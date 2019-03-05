package console_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var jobConfigPath = "./../../../../prow/jobs/console/components/react-odata/react-odata.yaml"
var runIfChangedConfig = "^components/react-odata/"
var buildScriptPath = "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"
var ODataComponentPath = "/home/prow/go/src/github.com/kyma-project/console/components/react-odata"

func TestReactODataJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(jobConfigPath)
	// THEN
	require.NoError(t, err)

	expName := "pre-master-console-react-odata"
	actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/console"], expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"master"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/console", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetBuildPr)
	assert.Equal(t, runIfChangedConfig, actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "components/react-odata/some_random_file.js")
	assert.Equal(t, tester.ImageNodeBuildpackLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{buildScriptPath}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{ODataComponentPath}, actualPresubmit.Spec.Containers[0].Args)
}

func TestReactODataJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(jobConfigPath)
	// THEN
	require.NoError(t, err)

	expName := "post-master-console-react-odata"
	actualPost := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/console"], expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"master"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/console", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetBuildConsoleMaster)
	assert.Equal(t, runIfChangedConfig, actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageNodeBuildpackLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{buildScriptPath}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{ODataComponentPath}, actualPost.Spec.Containers[0].Args)
}
