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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	magnitv1alpha1 "github.com/rashaev/crontab-operator/api/v1alpha1"

	"github.com/rashaev/crontab-operator/assets"
	batchv1 "k8s.io/api/batch/v1"
)

// CronTabReconciler reconciles a CronTab object
type CronTabReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=magnit.magnit.com,resources=crontabs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=magnit.magnit.com,resources=crontabs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=magnit.magnit.com,resources=crontabs/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=cronjob,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CronTab object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *CronTabReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	operatorCR := &magnitv1alpha1.CronTab{}
	err := r.Get(ctx, req.NamespacedName, operatorCR)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Operator resource object not found")
		return ctrl.Result{}, nil
	} else if err != nil {
		logger.Error(err, "Error getting operator resource object")
		return ctrl.Result{}, err
	}

	cronjob := &batchv1.CronJob{}
	create := false
	err = r.Get(ctx, req.NamespacedName, cronjob)
	if err != nil && errors.IsNotFound(err) {
		create = true
		cronjob = assets.GetCronJobFromFile("assets/cronjob.yaml")
	} else if err != nil {
		logger.Error(err, "Error getting existing CronTab")
		return ctrl.Result{}, err
	}

	cronjob.Namespace = req.Namespace
	cronjob.Name = req.Name

	if operatorCR.Spec.Command != nil {
		cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Command = operatorCR.Spec.Command
	}
	ctrl.SetControllerReference(operatorCR, cronjob, r.Scheme)

	if create {
		err = r.Create(ctx, cronjob)
	} else {
		err = r.Update(ctx, cronjob)
	}
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *CronTabReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&magnitv1alpha1.CronTab{}).
		Owns(&batchv1.CronJob{}).
		Complete(r)
}
