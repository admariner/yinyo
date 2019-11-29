package jobdispatcher

import (
	"regexp"

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type kubernetesClient struct {
	clientset *kubernetes.Clientset
}

// TODO: Rename this to yinyo-runs? We're avoiding using the word scrapers elsewhere
const namespace = "yinyo-scrapers"

// NewKubernetes returns the Kubernetes implementation of Client
func NewKubernetes() (Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	k := &kubernetesClient{clientset: clientset}
	return k, nil
}

func (client *kubernetesClient) CreateJobAndToken(namePrefix string, runToken string) (string, error) {
	// We need to convert the user-supplied namePrefix to something that will
	// work in k8s. That means only alpha numeric characters and "-".
	// For instance no "/".

	// Matches any non alphanumeric character
	re := regexp.MustCompile(`[^[:alnum:]]`)
	convertedNamePrefix := re.ReplaceAllString(namePrefix, "-")

	secretsClient := client.clientset.CoreV1().Secrets(namespace)
	secret := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: convertedNamePrefix + "-",
		},
		StringData: map[string]string{
			"run_token": runToken,
		},
	}
	created, err := secretsClient.Create(secret)

	return created.ObjectMeta.Name, err
}

func (client *kubernetesClient) StartJob(runName string, dockerImage string, command []string) error {
	jobsClient := client.clientset.BatchV1().Jobs(namespace)

	autoMountServiceAccountToken := false
	// Allow the job to get restarted up to 5 times before it's considered failed
	backOffLimit := int32(5)
	// Let this run for a maximum of 24 hours
	activeDeadlineSeconds := int64(86400)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: runName,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:          &backOffLimit,
			ActiveDeadlineSeconds: &activeDeadlineSeconds,
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					AutomountServiceAccountToken: &autoMountServiceAccountToken,
					RestartPolicy:                "OnFailure",
					Containers: []apiv1.Container{
						{
							Name:    runName,
							Image:   dockerImage,
							Command: command,
						},
					},
				},
			},
		},
	}
	_, err := jobsClient.Create(job)
	return err
}

func (client *kubernetesClient) DeleteJobAndToken(runName string) error {
	err := deleteJob(client.clientset, runName)
	if err != nil {
		return err
	}
	return deleteSecret(client.clientset, runName)
}

func deleteJob(clientset *kubernetes.Clientset, runName string) error {
	jobsClient := clientset.BatchV1().Jobs(namespace)

	deletePolicy := metav1.DeletePropagationForeground
	err := jobsClient.Delete(runName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		// Don't error if it's just that the job doesn't exist
		if err.(*apierrors.StatusError).ErrStatus.Reason == metav1.StatusReasonNotFound {
			return nil
		}
		return err
	}
	return nil
}

func deleteSecret(clientset *kubernetes.Clientset, runName string) error {
	secretsClient := clientset.CoreV1().Secrets(namespace)
	deletePolicy := metav1.DeletePropagationForeground
	err := secretsClient.Delete(runName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	return err
}

func (client *kubernetesClient) GetToken(runName string) (string, error) {
	// First get the actual run token from the secret
	secretsClient := client.clientset.CoreV1().Secrets(namespace)
	secret, err := secretsClient.Get(runName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	actualRunToken := string(secret.Data["run_token"])
	return actualRunToken, nil
}
