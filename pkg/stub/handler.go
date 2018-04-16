package stub

import (
	"fmt"
	"reflect"

	"github.com/banzaicloud/infinispan-operator/pkg/apis/infinispan/v1alpha1"
	"github.com/coreos/operator-sdk/pkg/sdk/action"
	"github.com/coreos/operator-sdk/pkg/sdk/handler"
	"github.com/coreos/operator-sdk/pkg/sdk/query"
	"github.com/coreos/operator-sdk/pkg/sdk/types"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

func NewHandler() handler.Handler {
	return &Handler{}
}

type Handler struct {
}

func (h *Handler) Handle(ctx types.Context, event types.Event) error {
	fmt.Printf("Handle: %+v %+v\n", event, event.Object)
	switch o := event.Object.(type) {
	case *v1alpha1.Infinispan:
		infinispan := o

		// Ignore the delete event since the garbage collector will clean up all secondary resources for the CR
		// All secondary resources must have the CR set as their OwnerReference for this to be the case
		if event.Deleted {
			return nil
		}

		// Create the deployment if it doesn't exist
		dep := deploymentForInfinispan(infinispan)
		err := action.Create(dep)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create deployment: %v", err)
		}

		// Ensure the deployment size is the same as the spec
		err = query.Get(dep)
		if err != nil {
			return fmt.Errorf("failed to get deployment: %v", err)
		}
		size := infinispan.Spec.Size
		if *dep.Spec.Replicas != size {
			dep.Spec.Replicas = &size
			err = action.Update(dep)
			if err != nil {
				return fmt.Errorf("failed to update deployment: %v", err)
			}
		}

		// Update the Infinispan status with the pod names
		podList := podList()
		labelSelector := labels.SelectorFromSet(labelsForInfinispan(infinispan.Name)).String()
		listOps := &metav1.ListOptions{LabelSelector: labelSelector}
		err = query.List(infinispan.Namespace, podList, query.WithListOptions(listOps))
		if err != nil {
			return fmt.Errorf("failed to list pods: %v", err)
		}
		podNames := getPodNames(podList.Items)
		if !reflect.DeepEqual(podNames, infinispan.Status.Nodes) {
			infinispan.Status.Nodes = podNames
			err := action.Update(infinispan)
			if err != nil {
				return fmt.Errorf("failed to update infinispan status: %v", err)
			}
		}

		// Create the Service for Infinispan
		ser := serviceForInfinispan(infinispan)
		err = action.Create(ser)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create service: %v", err)
		}
	}
	return nil
}

// deploymentForInfinispan returns a infinispan Deployment object
func deploymentForInfinispan(i *v1alpha1.Infinispan) *appsv1.Deployment {
	ls := labelsForInfinispan(i.Name)
	replicas := i.Spec.Size
	maxUnavailable := intstr.FromInt(1)
	maxSurge := intstr.FromInt(1)
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      i.Name,
			Namespace: i.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &maxUnavailable,
					MaxSurge:       &maxSurge,
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Image: "jboss/infinispan-server:latest",
						Name:  "infinispan-server",
						Ports: []v1.ContainerPort{
							{
								ContainerPort: 8181,
								Name:          "websocket",
							},
							{
								ContainerPort: 9990,
								Name:          "management",
							},
							{
								ContainerPort: 11211,
								Name:          "memcached",
							},
							{
								ContainerPort: 11222,
								Name:          "hotrod",
							},
							{
								ContainerPort: 7600,
								Name:          "jgroups",
							},
							{
								ContainerPort: 57600,
								Name:          "jgroups-fd",
							},
							{
								ContainerPort: 8080,
								Name:          "rest",
							},
						},
						Env: []v1.EnvVar{
							{
								Name: "KUBERNETES_NAMESPACE",
								ValueFrom: &v1.EnvVarSource{
									FieldRef: &v1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.namespace",
									},
								},
							},
							{
								Name:  "APP_USER",
								Value: "user",
							},
							{
								Name:  "APP_PASS",
								Value: "changeme",
							},
							{
								Name:  "MGMT_USER",
								Value: "admin",
							},
							{
								Name:  "MGMT_PASS",
								Value: "admin",
							},
						},
						LivenessProbe: &v1.Probe{
							Handler: v1.Handler{
								Exec: &v1.ExecAction{
									Command: []string{"/usr/local/bin/is_running.sh"},
								}},
							InitialDelaySeconds: 10,
							TimeoutSeconds:      80,
							PeriodSeconds:       60,
							SuccessThreshold:    1,
							FailureThreshold:    5,
						},
						ReadinessProbe: &v1.Probe{
							Handler: v1.Handler{
								Exec: &v1.ExecAction{
									Command: []string{"/usr/local/bin/is_healthy.sh"},
								}},
							InitialDelaySeconds: 10,
							TimeoutSeconds:      40,
							PeriodSeconds:       30,
							SuccessThreshold:    2,
							FailureThreshold:    5,
						},
					}},
				},
			},
		},
	}
	addOwnerRefToObject(dep, asOwner(i))
	return dep
}

// labelsForInfinispan returns the labels for selecting the resources
// belonging to the given infinispan CR name.
func labelsForInfinispan(name string) map[string]string {
	return map[string]string{"app": "infinispan", "infinispan_cr": name}
}

// addOwnerRefToObject appends the desired OwnerReference to the object
func addOwnerRefToObject(obj metav1.Object, ownerRef metav1.OwnerReference) {
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

// asOwner returns an OwnerReference set as the infinispan CR
func asOwner(i *v1alpha1.Infinispan) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: i.APIVersion,
		Kind:       i.Kind,
		Name:       i.Name,
		UID:        i.UID,
		Controller: &trueVar,
	}
}

// podList returns a v1.PodList object
func podList() *v1.PodList {
	return &v1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []v1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func serviceForInfinispan(i *v1alpha1.Infinispan) *v1.Service {
	ls := labelsForInfinispan(i.Name)
	service := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      i.Name,
			Namespace: i.Namespace,
		},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceTypeNodePort,
			Selector: ls,
			Ports: []v1.ServicePort{
				{
					Name: "rest",
					Port: 8080,
				},
				{
					Name: "management",
					Port: 9990,
				},
			},
		},
	}
	addOwnerRefToObject(service, asOwner(i))
	return service
}
