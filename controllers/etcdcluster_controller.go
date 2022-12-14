/*
Copyright 2022 cnych.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	etcdv1alpha1 "etcd-operator/api/v1alpha1"
)

// EtcdClusterReconciler reconciles a EtcdCluster object
type EtcdClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=etcd.ydzs.io,resources=etcdclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=etcd.ydzs.io,resources=etcdclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete

// controllers/etcdcluster_controller.go

func (r *EtcdClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("etcdcluster", req.NamespacedName)

	// 首先我们获取 EtcdCluster 实例
	var etcdCluster etcdv1alpha1.EtcdCluster
	if err := r.Get(ctx, req.NamespacedName, &etcdCluster); err != nil {
		// EtcdCluster was deleted，Ignore
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 得到 EtcdCluster 过后去创建对应的StatefulSet和Service
	// CreateOrUpdate

	// (就是观察的当前状态和期望的状态进行对比)

	// 调谐，获取到当前的一个状态，然后和我们期望的状态进行对比是不是就可以

	// CreateOrUpdate Service
	var svc corev1.Service
	svc.Name = etcdCluster.Name
	svc.Namespace = etcdCluster.Namespace

	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		or, err := ctrl.CreateOrUpdate(ctx, r, &svc, func() error {
			// 调谐必须在这个函数中去实现
			MutateHeadlessSvc(&etcdCluster, &svc)
			return controllerutil.SetControllerReference(&etcdCluster, &svc, r.Scheme)
		})
		log.Info("CreateOrUpdate", "Service", or)
		return err
	}); err != nil {
		return ctrl.Result{}, err
	}

	// CreateOrUpdate StatefulSet
	var sts appsv1.StatefulSet
	sts.Name = etcdCluster.Name
	sts.Namespace = etcdCluster.Namespace

	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		or, err := ctrl.CreateOrUpdate(ctx, r, &sts, func() error {
			// 调谐必须在这个函数中去实现
			MutateStatefulSet(&etcdCluster, &sts)
			return controllerutil.SetControllerReference(&etcdCluster, &sts, r.Scheme)
		})
		log.Info("CreateOrUpdate", "StatefulSet", or)
		return err
	}); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *EtcdClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&etcdv1alpha1.EtcdCluster{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
