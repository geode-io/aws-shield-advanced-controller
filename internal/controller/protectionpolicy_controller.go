package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	shieldawsv1alpha1 "github.com/geode-io/aws-shield-advanced-controller/api/v1alpha1"

	"github.com/geode-io/aws-shield-advanced-controller/internal/aws"
)

// ProtectionPolicyReconciler reconciles a ProtectionPolicy object
type ProtectionPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	ShieldManager aws.ShieldManager
}

//+kubebuilder:rbac:groups=shield.aws.geode.io,resources=protectionpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=shield.aws.geode.io,resources=protectionpolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=shield.aws.geode.io,resources=protectionpolicies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ProtectionPolicy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.2/pkg/reconcile
func (r *ProtectionPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("reconciling ProtectionPolicy", "req", req)

	var policy shieldawsv1alpha1.ProtectionPolicy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		log.Error(err, "unable to fetch ProtectionPolicy")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProtectionPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&shieldawsv1alpha1.ProtectionPolicy{}).
		Complete(r)
}
