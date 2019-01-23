package helm

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gruntwork-io/kubergrunt/kubectl"
	"github.com/gruntwork-io/kubergrunt/tls"
)

func TestStoreCertificateKeyPairAsKubernetesSecretStoresAllFiles(t *testing.T) {
	t.Parallel()

	// Construct kubectl options
	ttKubectlOptions := k8s.NewKubectlOptions("", "")
	configPath, err := k8s.KubeConfigPathFromHomeDirE()
	require.NoError(t, err)
	kubectlOptions := kubectl.NewKubectlOptions("", configPath)

	// Create a namespace so we don't collide with other tests
	namespace := strings.ToLower(random.UniqueId())
	k8s.CreateNamespace(t, ttKubectlOptions, namespace)
	defer k8s.DeleteNamespace(t, ttKubectlOptions, namespace)

	// Now store certificate key pair using the tested function
	baseName := random.UniqueId()
	certificateKeyPairPath := createSampleCertificateKeyPairPath(t)
	err = StoreCertificateKeyPairAsKubernetesSecret(
		kubectlOptions,
		"random-certs",
		namespace,
		map[string]string{},
		map[string]string{},
		baseName,
		certificateKeyPairPath,
	)
	require.NoError(t, err)

	// Verify the created cert
	ttKubectlOptions.Namespace = namespace
	secret := k8s.GetSecret(t, ttKubectlOptions, "random-certs")
	assert.Equal(t, secret.Data[fmt.Sprintf("%s.crt", baseName)], mustReadFile(t, certificateKeyPairPath.CertificatePath))
	assert.Equal(t, secret.Data[fmt.Sprintf("%s.pem", baseName)], mustReadFile(t, certificateKeyPairPath.PrivateKeyPath))
	assert.Equal(t, secret.Data[fmt.Sprintf("%s.pub", baseName)], mustReadFile(t, certificateKeyPairPath.PublicKeyPath))
}

func mustReadFile(t *testing.T, path string) []byte {
	data, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	return data
}

func createSampleCertificateKeyPairPath(t *testing.T) tls.CertificateKeyPairPath {
	return tls.CertificateKeyPairPath{
		CertificatePath: mustAbs(t, "./testfixtures/ca.cert"),
		PrivateKeyPath:  mustAbs(t, "./testfixtures/tls.pem"),
		PublicKeyPath:   mustAbs(t, "./testfixtures/tls.pub"),
	}
}

func mustAbs(t *testing.T, path string) string {
	absPath, err := filepath.Abs(path)
	require.NoError(t, err)
	return absPath
}
