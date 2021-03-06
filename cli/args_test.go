package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terragrunt/config"
	"github.com/gruntwork-io/terragrunt/errors"
	"github.com/gruntwork-io/terragrunt/options"
	"github.com/gruntwork-io/terragrunt/util"
	"github.com/stretchr/testify/assert"
)

func TestParseTerragruntOptionsFromArgs(t *testing.T) {
	t.Parallel()

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	workingDir = filepath.ToSlash(workingDir)

	testCases := []struct {
		args            []string
		expectedOptions *options.TerragruntOptions
		expectedErr     error
	}{
		{
			[]string{},
			mockOptions(util.JoinPath(workingDir, config.DefaultTerragruntConfigPath), workingDir, []string{}, false, "", false),
			nil,
		},

		{
			[]string{"foo", "bar"},
			mockOptions(util.JoinPath(workingDir, config.DefaultTerragruntConfigPath), workingDir, []string{"foo", "bar"}, false, "", false),
			nil,
		},

		{
			[]string{"--foo", "--bar"},
			mockOptions(util.JoinPath(workingDir, config.DefaultTerragruntConfigPath), workingDir, []string{"--foo", "--bar"}, false, "", false),
			nil,
		},

		{
			[]string{"--foo", "apply", "--bar"},
			mockOptions(util.JoinPath(workingDir, config.DefaultTerragruntConfigPath), workingDir, []string{"--foo", "apply", "--bar"}, false, "", false),
			nil,
		},

		{
			[]string{"--terragrunt-non-interactive"},
			mockOptions(util.JoinPath(workingDir, config.DefaultTerragruntConfigPath), workingDir, []string{}, true, "", false),
			nil,
		},

		{
			[]string{"--terragrunt-config", fmt.Sprintf("/some/path/%s", config.DefaultTerragruntConfigPath)},
			mockOptions(fmt.Sprintf("/some/path/%s", config.DefaultTerragruntConfigPath), workingDir, []string{}, false, "", false),
			nil,
		},

		{
			[]string{"--terragrunt-working-dir", "/some/path"},
			mockOptions(util.JoinPath("/some/path", config.DefaultTerragruntConfigPath), "/some/path", []string{}, false, "", false),
			nil,
		},

		{
			[]string{"--terragrunt-source", "/some/path"},
			mockOptions(util.JoinPath(workingDir, config.DefaultTerragruntConfigPath), workingDir, []string{}, false, "/some/path", false),
			nil,
		},

		{
			[]string{"--terragrunt-ignore-dependency-errors"},
			mockOptions(util.JoinPath(workingDir, config.DefaultTerragruntConfigPath), workingDir, []string{}, false, "", true),
			nil,
		},

		{
			[]string{"--terragrunt-config", fmt.Sprintf("/some/path/%s", config.DefaultTerragruntConfigPath), "--terragrunt-non-interactive"},
			mockOptions(fmt.Sprintf("/some/path/%s", config.DefaultTerragruntConfigPath), workingDir, []string{}, true, "", false),
			nil,
		},

		{
			[]string{"--foo", "--terragrunt-config", fmt.Sprintf("/some/path/%s", config.DefaultTerragruntConfigPath), "bar", "--terragrunt-non-interactive", "--baz", "--terragrunt-working-dir", "/some/path", "--terragrunt-source", "github.com/foo/bar//baz?ref=1.0.3"},
			mockOptions(fmt.Sprintf("/some/path/%s", config.DefaultTerragruntConfigPath), "/some/path", []string{"--foo", "bar", "--baz"}, true, "github.com/foo/bar//baz?ref=1.0.3", false),
			nil,
		},

		{
			[]string{"--terragrunt-config"},
			nil,
			ArgMissingValue("terragrunt-config"),
		},

		{
			[]string{"--terragrunt-working-dir"},
			nil,
			ArgMissingValue("terragrunt-working-dir"),
		},

		{
			[]string{"--foo", "bar", "--terragrunt-config"},
			nil,
			ArgMissingValue("terragrunt-config"),
		},
	}

	for _, testCase := range testCases {
		actualOptions, actualErr := parseTerragruntOptionsFromArgs(testCase.args, &bytes.Buffer{}, &bytes.Buffer{})
		if testCase.expectedErr != nil {
			assert.True(t, errors.IsError(actualErr, testCase.expectedErr), "Expected error %v but got error %v", testCase.expectedErr, actualErr)
		} else {
			assert.Nil(t, actualErr, "Unexpected error: %v", actualErr)
			assertOptionsEqual(t, *testCase.expectedOptions, *actualOptions, "For args %v", testCase.args)
		}
	}
}

// We can't do a direct comparison between TerragruntOptions objects because we can't compare Logger or RunTerragrunt
// instances. Therefore, we have to manually check everything else.
func assertOptionsEqual(t *testing.T, expected options.TerragruntOptions, actual options.TerragruntOptions, msgAndArgs ...interface{}) {
	assert.NotNil(t, expected.Logger, msgAndArgs...)
	assert.NotNil(t, actual.Logger, msgAndArgs...)

	assert.Equal(t, expected.TerragruntConfigPath, actual.TerragruntConfigPath, msgAndArgs...)
	assert.Equal(t, expected.NonInteractive, actual.NonInteractive, msgAndArgs...)
	assert.Equal(t, expected.TerraformCliArgs, actual.TerraformCliArgs, msgAndArgs...)
	assert.Equal(t, expected.WorkingDir, actual.WorkingDir, msgAndArgs...)
	assert.Equal(t, expected.Source, actual.Source, msgAndArgs...)
	assert.Equal(t, expected.IgnoreDependencyErrors, actual.IgnoreDependencyErrors, msgAndArgs...)
}

func mockOptions(terragruntConfigPath string, workingDir string, terraformCliArgs []string, nonInteractive bool, terragruntSource string, ignoreDependencyErrors bool) *options.TerragruntOptions {
	opts := options.NewTerragruntOptionsForTest(terragruntConfigPath)

	opts.WorkingDir = workingDir
	opts.TerraformCliArgs = terraformCliArgs
	opts.NonInteractive = nonInteractive
	opts.Source = terragruntSource
	opts.IgnoreDependencyErrors = ignoreDependencyErrors

	return opts
}

func TestFilterTerragruntArgs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		args     []string
		expected []string
	}{
		{[]string{}, []string{}},
		{[]string{"foo", "--bar"}, []string{"foo", "--bar"}},
		{[]string{"foo", "--terragrunt-config", fmt.Sprintf("/some/path/%s", config.DefaultTerragruntConfigPath)}, []string{"foo"}},
		{[]string{"foo", "--terragrunt-non-interactive"}, []string{"foo"}},
		{[]string{"foo", "--terragrunt-non-interactive", "--bar", "--terragrunt-working-dir", "/some/path", "--baz", "--terragrunt-config", fmt.Sprintf("/some/path/%s", config.DefaultTerragruntConfigPath)}, []string{"foo", "--bar", "--baz"}},
		{[]string{"apply-all", "foo", "bar"}, []string{"foo", "bar"}},
		{[]string{"foo", "destroy-all", "--foo", "--bar"}, []string{"foo", "--foo", "--bar"}},
	}

	for _, testCase := range testCases {
		actual := filterTerragruntArgs(testCase.args)
		assert.Equal(t, testCase.expected, actual, "For args %v", testCase.args)
	}
}

func TestParseEnvironmentVariables(t *testing.T) {
	testCases := []struct {
		environmentVariables []string
		expectedVariables    map[string]string
	}{
		{
			[]string{},
			map[string]string{},
		},
		{
			[]string{"foobar"},
			map[string]string{},
		},
		{
			[]string{"foo=bar"},
			map[string]string{"foo": "bar"},
		},
		{
			[]string{"foo=bar", "goo=gar"},
			map[string]string{"foo": "bar", "goo": "gar"},
		},
		{
			[]string{"foo=bar   "},
			map[string]string{"foo": "bar   "},
		},
		{
			[]string{"foo   =bar   "},
			map[string]string{"foo": "bar   "},
		},
		{
			[]string{"foo=composite=bar"},
			map[string]string{"foo": "composite=bar"},
		},
	}

	for _, testCase := range testCases {
		actualVariables := parseEnvironmentVariables(testCase.environmentVariables)
		assert.Equal(t, testCase.expectedVariables, actualVariables)
	}
}
