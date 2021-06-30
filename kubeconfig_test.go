package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

func CleanUpKubernetesTest() {
	os.Rename("./test/config1.yml.backup", "./test/config1.yml")
	os.Rename("./test/config2.yml.backup", "./test/config2.yml")
}

func TestUpdateGlobalConfig(t *testing.T) {
	t.Cleanup(CleanUpKubernetesTest)
	LogConfig.Level.SetLevel(zap.DebugLevel)
	Sugar.Info("Test: UpdateGlobalConfig")

	var k8sconfig KubectlConfig
	content, err := ioutil.ReadFile("test/config0.yml")
	require.NoError(t, err)
	err = yaml.Unmarshal([]byte(content), &k8sconfig)
	require.NoError(t, err)
	require.NotNil(t, k8sconfig)

	conf := UpdateGlobalConfig("./test/config1.yml", &k8sconfig)
	require.NotEmpty(t, conf)

	added := false
	for _, context := range conf.Contexts {
		if context.Name == k8sconfig.CurrentContext {
			added = true
		}
	}
	assert.True(t, added)

	conf = UpdateGlobalConfig("./test/config2.yml", &k8sconfig)
	require.NotEmpty(t, conf)

	changed := false
	for _, context := range conf.Contexts {
		if context.Name == k8sconfig.CurrentContext && context.Context.User == k8sconfig.Contexts[0].Context.User {
			changed = true
		}
	}
	assert.True(t, changed)
}
