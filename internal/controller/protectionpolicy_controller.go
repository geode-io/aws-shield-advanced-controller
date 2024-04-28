package controller

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	shieldawsv1alpha1 "github.com/geode-io/aws-shield-advanced-controller/api/v1alpha1"

	"github.com/geode-io/aws-shield-advanced-controller/internal/aws"
	"github.com/geode-io/aws-shield-advanced-controller/internal/config"
)

// ProtectionPolicyReconciler reconciles a ProtectionPolicy object
type ProtectionPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Config        *config.Config
	ShieldManager aws.ShieldManager
	Discovery     aws.DiscoveryClient
}

//+kubebuilder:rbac:groups=shield.aws.geode.io,resources=protectionpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=shield.aws.geode.io,resources=protectionpolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=shield.aws.geode.io,resources=protectionpolicies/finalizers,verbs=update

func (r *ProtectionPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the ProtectionPolicy resource
	policy := &shieldawsv1alpha1.ProtectionPolicy{}
	if err := r.Get(ctx, req.NamespacedName, policy); err != nil {
		if apierrors.IsNotFound(err) {
			// ProtectionPolicy resource not found, no need to requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object, requeue the request
		return ctrl.Result{}, err
	}

	// Add the finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(policy, FinalizerName) {
		controllerutil.AddFinalizer(policy, FinalizerName)
		err := r.Update(ctx, policy)
		if err != nil {
			log.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
	}

	// Check if the Protection instance is marked for deletion
	if policy.GetDeletionTimestamp() != nil {
		// Protection is marked for deletion
		if controllerutil.ContainsFinalizer(policy, FinalizerName) {

			// Delete the protection resources in AWS
			if !r.Config.DryRun {
				protections := policy.Status.Protections
				for _, protection := range protections {
					err := r.ShieldManager.DeleteProtection(ctx, protection.ProtectionArn)
					if err != nil {
						log.Error(err, "Failed to delete protection", "protection", protection.ProtectionArn)
						return ctrl.Result{}, err
					}
				}
			} else {
				log.Info("Dry-run mode enabled, skipping deletion of protection resources")
			}

			// Remove the finalizer
			controllerutil.RemoveFinalizer(policy, FinalizerName)
			err := r.Update(ctx, policy)
			if err != nil {
				log.Error(err, "Failed to remove finalizer")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Find all resources that match the ProtectionPolicy
	resourcesTypes := []string{}
	for _, typ := range policy.Spec.MatchResourceTypes {
		resourcesTypes = append(resourcesTypes, string(typ))
	}

	resources, err := r.Discovery.Discover(ctx, &aws.DiscoveryRequest{
		ResourceTypes: resourcesTypes,
		Regions:       policy.Spec.MatchRegions,
	})
	if err != nil {
		log.Error(err, "Failed to discover resources")
		return ctrl.Result{}, err
	}

	log.Info("Discovered resources", "count", len(resources.Resources))
	log.V(1).Info("Discovered resources", "resources", resources.Resources)

	if r.Config.DryRun {
		log.Info("Dry-run mode enabled, skipping creation or update of protection resources")
		return ctrl.Result{}, nil
	}

	// Create or update protection resources in AWS and update status
	policy.Status.Protections = []shieldawsv1alpha1.ProtectionStatus{}
	for _, resource := range resources.Resources {
		protectionArn, err := r.ShieldManager.CreateOrUpdateProtection(ctx, resource.Name, resource.Arn)
		if err != nil {
			log.Error(err, "Failed to create protection", "resource", resource.Arn)
			return ctrl.Result{}, err
		}

		policy.Status.Protections = append(policy.Status.Protections, shieldawsv1alpha1.ProtectionStatus{
			State:         shieldawsv1alpha1.ProtectionStateActive,
			ProtectionArn: protectionArn,
			ResourceArn:   resource.Arn,
		})
	}

	// Write status
	err = r.Status().Update(ctx, policy)
	if err != nil {
		log.Error(err, "Failed to update ProtectionPolicy status")
		return ctrl.Result{}, err
	}

	// Requeue after the configured resync interval
	log.V(1).Info("Requeueing after resync interval", "interval", r.Config.PolicyResyncInterval)
	return ctrl.Result{
		RequeueAfter: r.Config.PolicyResyncInterval,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProtectionPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&shieldawsv1alpha1.ProtectionPolicy{}).
		Complete(r)
}
