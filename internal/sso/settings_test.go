package sso

/*
 * AWS SSO CLI
 * Copyright (c) 2021-2025 Aaron Turner  <synfinatic at gmail dot com>
 *
 * This program is free software: you can redistribute it
 * and/or modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or with the authors permission any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/synfinatic/aws-sso-cli/internal/url"
	"github.com/synfinatic/flexlog"
	testlogger "github.com/synfinatic/flexlog/test"
)

const (
	TEST_SETTINGS_FILE = "./testdata/settings.yaml"
)

var TEST_GET_ROLE_ARN []string = []string{
	"arn:aws:iam::258234615182:role/AWSAdministratorAccess",
	"arn:aws:iam::258234615182:role/LimitedAccess",
	"arn:aws:iam::833365043586:role/AWSAdministratorAccess",
}

type SettingsTestSuite struct {
	suite.Suite
	settings *Settings
}

func TestSettingsTestSuite(t *testing.T) {
	over := OverrideSettings{}
	defaults := map[string]interface{}{}
	settings, err := LoadSettings(TEST_SETTINGS_FILE, TEST_CACHE_FILE, defaults, over)
	assert.Nil(t, err)

	s := &SettingsTestSuite{
		settings: settings,
	}
	suite.Run(t, s)
}

func (suite *SettingsTestSuite) TestLoadSettings() {
	t := suite.T()

	assert.Equal(t, TEST_SETTINGS_FILE, suite.settings.ConfigFile())

	assert.Equal(t, "", suite.settings.ConfigUrlAction) // deprecated
	// ensure we upgraded ConfigUrlAction to UrlAction
	assert.Equal(t, url.OpenUrlContainer, suite.settings.UrlAction)
	// ensure we applied UrlAction to ConfigProfilesUrlAction
	assert.Equal(t, url.ConfigProfilesOpen, suite.settings.ConfigProfilesUrlAction)

	// ensure we upgraded FirefoxOpenUrlInContainer
	assert.False(t, suite.settings.FirefoxOpenUrlInContainer)
	assert.Equal(t, url.OpenUrlContainer, suite.settings.UrlAction)
}

func TestConfigProfilesUrlAction(t *testing.T) {
	t.Parallel()
	settingsFile := "./testdata/settings2.yaml"
	settings, err := LoadSettings(settingsFile, TEST_CACHE_FILE, map[string]interface{}{}, OverrideSettings{})
	assert.NoError(t, err)

	assert.Equal(t, url.ConfigProfilesOpen, settings.ConfigProfilesUrlAction)

	settingsFile = "./testdata/settings3.yaml"
	settings, err = LoadSettings(settingsFile, TEST_CACHE_FILE, map[string]interface{}{}, OverrideSettings{})
	assert.NoError(t, err)

	assert.Equal(t, url.ConfigProfilesOpenUrlContainer, settings.ConfigProfilesUrlAction)
}

func (suite *SettingsTestSuite) TestGetSelectedSSO() {
	t := suite.T()

	sso, err := suite.settings.GetSelectedSSO("Default")
	assert.NoError(t, err)
	assert.Equal(t, "https://d-754545454.awsapps.com/start", sso.StartUrl)

	sso, err = suite.settings.GetSelectedSSO("Another")
	assert.NoError(t, err)
	assert.Equal(t, "https://d-755555555.awsapps.com/start", sso.StartUrl)

	sso, err = suite.settings.GetSelectedSSO("Bug292")
	assert.NoError(t, err)
	assert.Equal(t, "https://d-88888888888.awsapps.com/start", sso.StartUrl)

	sso, err = suite.settings.GetSelectedSSO("Foobar")
	assert.Error(t, err)
	assert.Equal(t, "", sso.StartUrl)

	sso, err = suite.settings.GetSelectedSSO("")
	assert.NoError(t, err)
	assert.Equal(t, "https://d-754545454.awsapps.com/start", sso.StartUrl)
}

func (suite *SettingsTestSuite) TestGetSelectedSSOName() {
	t := suite.T()

	name, err := suite.settings.GetSelectedSSOName("Default")
	assert.NoError(t, err)
	assert.Equal(t, "Default", name)

	name, err = suite.settings.GetSelectedSSOName("Foobar")
	assert.Error(t, err)
	assert.Equal(t, "", name)

	name, err = suite.settings.GetSelectedSSOName("Another")
	assert.NoError(t, err)
	assert.Equal(t, "Another", name)

	name, err = suite.settings.GetSelectedSSOName("")
	assert.NoError(t, err)
	assert.Equal(t, "Default", name)

	// what if user removes this value from the config file?
	s := *suite.settings
	s.DefaultSSO = ""
	name, err = (&s).GetSelectedSSOName("")
	assert.NoError(t, err)
	assert.Equal(t, "Default", name)
}

func (suite *SettingsTestSuite) TestCreatedAt() {
	t := suite.T()
	sso, _ := suite.settings.GetSelectedSSO("")
	assert.Equal(t, sso.CreatedAt(), suite.settings.CreatedAt())
}

func (suite *SettingsTestSuite) TestGetRoles() {
	t := suite.T()

	sso, _ := suite.settings.GetSelectedSSO("")
	roles := sso.GetRoles()

	// makes sure we found the 3 roles...
	assert.Equal(t, 3, len(roles))

	// and their ARN's match
	arns := []string{}
	for _, role := range roles {
		arns = append(arns, role.ARN)
	}
	for _, role := range TEST_GET_ROLE_ARN {
		assert.Contains(t, arns, role)
	}

	// make sure we can parse this yaml
	sso, _ = suite.settings.GetSelectedSSO("Bug292")
	roles = sso.GetRoles()
	assert.Equal(t, 1, len(roles))
}

func (suite *SettingsTestSuite) TestGetAllTags() {
	t := suite.T()

	sso, _ := suite.settings.GetSelectedSSO("")
	tagsPtr := sso.GetAllTags()
	tags := *tagsPtr
	assert.ElementsMatch(t, tags["Test"], []string{"value", "logs"})
	assert.ElementsMatch(t, tags["Foo"], []string{"Bar", "Moo"})
	assert.ElementsMatch(t, tags["Can"], []string{"Man"})
	assert.ElementsMatch(t, tags["DoesNotExistTag"], []string{})
}

func (suite *SettingsTestSuite) TestSave() {
	t := suite.T()

	dir, err := os.MkdirTemp("", "settings_test")
	assert.Nil(t, err)
	defer os.RemoveAll(dir)

	p := filepath.Join(dir, "foo/bar/config.yaml")
	err = suite.settings.Save(p, true)
	assert.Nil(t, err)
	err = suite.settings.Save(p, false)
	assert.NotNil(t, err)

	err = suite.settings.Save(dir, false)
	assert.NotNil(t, err)
}

func clearEnv() {
	os.Setenv("AWS_DEFAULT_REGION", "")
	os.Setenv("AWS_SSO_DEFAULT_REGION", "")
}

func (suite *SettingsTestSuite) TestGetDefaultRegion() {
	t := suite.T()

	clearEnv()
	defer clearEnv()

	assert.Equal(t, "ca-central-1", suite.settings.GetDefaultRegion(258234615182, "AWSAdministratorAccess", false))
	assert.Equal(t, "eu-west-1", suite.settings.GetDefaultRegion(258234615182, "LimitedAccess", false))
	assert.Equal(t, "us-east-1", suite.settings.GetDefaultRegion(833365043586, "AWSAdministratorAccess:", false))

	assert.Equal(t, "", suite.settings.GetDefaultRegion(258234615182, "AWSAdministratorAccess", true))
	assert.Equal(t, "", suite.settings.GetDefaultRegion(258234615182, "LimitedAccess", true))
	assert.Equal(t, "", suite.settings.GetDefaultRegion(833365043586, "AWSAdministratorAccess:", true))

	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	assert.Equal(t, "", suite.settings.GetDefaultRegion(258234615182, "AWSAdministratorAccess", false))
	assert.Equal(t, "", suite.settings.GetDefaultRegion(258234615182, "LimitedAccess", false))
	assert.Equal(t, "", suite.settings.GetDefaultRegion(833365043586, "AWSAdministratorAccess:", false))

	os.Setenv("AWS_SSO_DEFAULT_REGION", "us-east-1")
	assert.Equal(t, "ca-central-1", suite.settings.GetDefaultRegion(258234615182, "AWSAdministratorAccess", false))
	assert.Equal(t, "eu-west-1", suite.settings.GetDefaultRegion(258234615182, "LimitedAccess", false))
	assert.Equal(t, "us-east-1", suite.settings.GetDefaultRegion(833365043586, "AWSAdministratorAccess:", false))

	oldLogger := log.Copy()
	tLogger := testlogger.NewTestLogger("DEBUG")
	defer tLogger.Close()
	log = tLogger
	defer func() { log = oldLogger }()

	suite.settings.GetDefaultRegion(-1, "foo", false)
	msg := testlogger.LogMessage{}
	assert.NoError(t, tLogger.GetNext(&msg))
	assert.Contains(t, msg.Message, "Unable to GetDefaultRegion")
	assert.Equal(t, flexlog.LevelFatal, msg.Level)
}

func (suite *SettingsTestSuite) TestOtherSSO() {
	t := suite.T()
	over := OverrideSettings{
		DefaultSSO: "Another",
	}
	defaults := map[string]interface{}{}
	settings, err := LoadSettings(TEST_SETTINGS_FILE, TEST_CACHE_FILE, defaults, over)
	assert.Nil(t, err)

	assert.Equal(t, "us-west-2", settings.GetDefaultRegion(182347455, "AWSAdministratorAccess", false))
}

func (suite *SettingsTestSuite) TestGetEnvVarTags() {
	t := suite.T()

	x := map[string]string{
		"Role": "AWS_SSO_TAG_ROLE",
		"Arn":  "AWS_SSO_TAG_ARN",
		"Foo":  "AWS_SSO_TAG_FOO",
	}
	y := suite.settings.GetEnvVarTags()
	assert.EqualValues(t, x, y)
}

func (suite *SettingsTestSuite) TestGetAllProfiles() {
	t := suite.T()

	getExecutable = func() (string, error) { return "", fmt.Errorf("failed") }
	_, err := suite.settings.GetAllProfiles()
	assert.Error(t, err)

	getExecutable = os.Executable

	profiles, err := suite.settings.GetAllProfiles()
	assert.NoError(t, err)

	assert.Len(t, *profiles, 1)
	assert.Contains(t, *profiles, "Default")

	p := *profiles
	d := p["Default"]
	assert.Len(t, d, 19)

	x, ok := d["arn:aws:iam::833365043586:role/AWSAdministratorAccess"]
	assert.True(t, ok)

	assert.Equal(t, x.Arn, "arn:aws:iam::833365043586:role/AWSAdministratorAccess")
	assert.NotEmpty(t, x.BinaryPath)
	assert.Equal(t, map[string]interface{}(nil), x.ConfigVariables)
	assert.Equal(t, "Log archive/AWSAdministratorAccess", x.Profile)
	assert.Equal(t, "Default", x.Sso)

	assert.NoError(t, profiles.UniqueCheck(suite.settings))

	assert.False(t, profiles.IsDuplicate("testing"))
	assert.True(t, profiles.IsDuplicate("Log archive/AWSAdministratorAccess"))

	oldFormat := suite.settings.ProfileFormat
	// generates duplicates
	suite.settings.ProfileFormat = "{{ .AccountId }}"
	assert.Error(t, profiles.UniqueCheck(suite.settings))

	// unable to generate a profile
	suite.settings.ProfileFormat = "{{ .UniqueCheckFailure }}"
	assert.Error(t, profiles.UniqueCheck(suite.settings))

	suite.settings.ProfileFormat = "{{ .GetAllProfilesFailure }}"
	_, err = suite.settings.GetAllProfiles()
	assert.Error(t, err)

	suite.settings.ProfileFormat = oldFormat
}

func (suite *SettingsTestSuite) TestValidate() {
	t := suite.T()

	assert.NoError(t, suite.settings.Validate())
	suite.settings.UrlAction = url.Exec
	suite.settings.ConfigProfilesUrlAction = url.ConfigProfilesGrantedContainer
	assert.Error(t, suite.settings.Validate())
}

func (suite *SettingsTestSuite) TestSetOverrides() {
	t := suite.T()

	s := suite.settings
	overrides := OverrideSettings{
		LogLevel:   "debug",
		LogLines:   true,
		Browser:    "my-browser",
		DefaultSSO: "hello",
		Threads:    10,
	}

	s.setOverrides(overrides)

	assert.Equal(t, slog.LevelDebug, log.GetLevel())
	// assert.True(t, log.ReportCaller)
	assert.Equal(t, "my-browser", s.Browser)
	assert.Equal(t, "hello", s.DefaultSSO)
	assert.Equal(t, 10, s.Threads)
}

func (suite *SettingsTestSuite) TestGetSSOProfiles() {
	t := suite.T()

	profiles, err := suite.settings.GetSSOProfiles("Default")
	assert.NoError(t, err)
	assert.NotNil(t, profiles)
	assert.Equal(t, 1, len(*profiles))
	assert.Contains(t, *profiles, "Default")
	assert.NotContains(t, *profiles, "Another")
}

func TestCreatedAt(t *testing.T) {
	s := Settings{
		configFile: "/dev/null/invalid",
	}

	oldLogger := log.Copy()
	tLogger := testlogger.NewTestLogger("DEBUG")
	defer tLogger.Close()
	log = tLogger
	defer func() { log = oldLogger }()

	assert.Panics(t, func() { s.CreatedAt() }) // will panic because log.Fatal() doesn't return
	msg := testlogger.LogMessage{}
	assert.NoError(t, tLogger.GetNext(&msg))
	assert.Contains(t, msg.Message, "Unable to open")
	assert.Equal(t, flexlog.LevelFatal, msg.Level)
}

func TestApplyDeprecations(t *testing.T) {
	s := &Settings{
		ListFields:                []string{"Foo", "Bar", "ExpiresStr", "AccountIdStr", "ARN"},
		ProfileFormat:             "{{ AccountIdStr .AccountId }}:{{ .RoleName }}",
		FirefoxOpenUrlInContainer: true,
		ConfigProfilesUrlAction:   url.ConfigProfilesUndef,
		ConfigUrlAction:           string(url.Exec),
	}

	r := s.applyDeprecations()
	assert.True(t, r)

	// Upgrade ConfigUrlAction to ConfigProfilesUrlAction
	assert.Equal(t, string(url.Undef), s.ConfigUrlAction)
	assert.Equal(t, url.ConfigProfilesExec, s.ConfigProfilesUrlAction)

	// Upgrade FirefoxOpenUrlInContainer to UrlAction
	assert.Equal(t, url.OpenUrlContainer, s.UrlAction)
	assert.Equal(t, false, s.FirefoxOpenUrlInContainer)

	// ExpiresStr => Expires, etc
	assert.Equal(t, []string{"Foo", "Bar", "Expires", "AccountIdPad", "Arn"}, s.ListFields)

	// AccountIdStr .AccountId => .AccountIdPad
	assert.Equal(t, "{{ .AccountIdPad }}:{{ .RoleName }}", s.ProfileFormat)
}

func TestGetExecutable(t *testing.T) {
	path, err := getExecutable()
	assert.NoError(t, err)
	// can't test the NIX path really can we??
	assert.Contains(t, path, "sso.test")
}
