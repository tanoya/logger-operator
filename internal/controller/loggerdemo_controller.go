/*
Copyright 2024.

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

package controller

import (
	"context"

	githubv1 "github.com/logger-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// LoggerDemoReconciler reconciles a LoggerDemo object
type LoggerDemoReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=github.github.com,resources=loggerdemoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=github.github.com,resources=loggerdemoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=github.github.com,resources=loggerdemoes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LoggerDemo object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *LoggerDemoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLog := log.FromContext(ctx)
	reqLog.Info("Reconciling LoggerDemo")

	// Fetch the LoggerDemo instance
	instance := &githubv1.LoggerDemo{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			reqLog.Info("LoggerDemo resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		reqLog.Error(err, "Failed to get LoggerDemo")
		return ctrl.Result{}, err
	}

	pod := newPodForCR(instance)
	if err := ctrl.SetControllerReference(instance, pod, r.Scheme); err != nil {
		reqLog.Error(err, "Failed to set controller reference")
		return ctrl.Result{}, err
	}

	found := &corev1.Pod{}
	err = r.Get(ctx, client.ObjectKey{Namespace: pod.Namespace, Name: pod.Name}, found)
	if err != nil && client.IgnoreNotFound(err) != nil {
		reqLog.Error(err, "Failed to get Pod")
		return ctrl.Result{}, err
	}
	reqLog.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
	err = r.Create(ctx, pod)
	if err != nil {
		reqLog.Error(err, "Failed to create new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		return ctrl.Result{}, err
	}
	reqLog.Info("Created a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name, "success.")

	return ctrl.Result{}, nil
}

func newPodForCR(cr *githubv1.LoggerDemo) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "logger-demo",
					Image: cr.Spec.Image,
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *LoggerDemoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&githubv1.LoggerDemo{}).
		Complete(r)
}
