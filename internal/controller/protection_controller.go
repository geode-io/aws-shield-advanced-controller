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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	shieldawsv1alpha1 "github.com/geode-io/aws-shield-advanced-controller/api/v1alpha1"

	"github.com/geode-io/aws-shield-advanced-controller/internal/aws"
)

// ProtectionReconciler reconciles a Protection object
type ProtectionReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	ShieldManager aws.ShieldManager
}

//+kubebuilder:rbac:groups=shield.aws.geode.io,resources=protections,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=shield.aws.geode.io,resources=protections/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=shield.aws.geode.io,resources=protections/finalizers,verbs=update

func (r *ProtectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the Protection instance
	protection := &shieldawsv1alpha1.Protection{}
	if err := r.Get(ctx, req.NamespacedName, protection); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if the Protection instance is marked for deletion
	if protection.GetDeletionTimestamp() != nil {
		// Protection is marked for deletion
		if controllerutil.ContainsFinalizer(protection, FinalizerName) {
			// Perform cleanup tasks if necessary
			err := r.ShieldManager.DeleteProtection(ctx, protection.Status.ProtectionArn)
			if err != nil {
				log.Error(err, "Failed to delete resource protection")
				return ctrl.Result{}, err
			}

			// Remove the finalizer
			controllerutil.RemoveFinalizer(protection, FinalizerName)
			err = r.Update(ctx, protection)
			if err != nil {
				log.Error(err, "Failed to remove finalizer")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add the finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(protection, FinalizerName) {
		controllerutil.AddFinalizer(protection, FinalizerName)
		err := r.Update(ctx, protection)
		if err != nil {
			log.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
	}

	// Create or update the resource protection in AWS Shield Advanced
	protectionArn, err := r.ShieldManager.CreateOrUpdateProtection(
		ctx,
		protection.Name,
		protection.Spec.ResourceArn,
	)
	if err != nil {
		log.Error(err, "Failed to create or update resource protection")
		return ctrl.Result{}, err
	}

	// Update the status of the Protection resource
	protection.Status.ProtectionArn = protectionArn
	protection.Status.State = shieldawsv1alpha1.ProtectionStateActive
	err = r.Status().Update(ctx, protection)
	if err != nil {
		log.Error(err, "Failed to update Protection status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProtectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&shieldawsv1alpha1.Protection{}).
		Complete(r)
}
