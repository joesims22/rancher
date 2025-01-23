package clusterandprojectroles

import (
	"testing"
	"fmt"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/rancher/tests/v2/actions/clusters"
	"github.com/rancher/rancher/tests/v2/actions/provisioning"
	"github.com/rancher/rancher/tests/v2/actions/provisioninginput"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	rbac "github.com/rancher/rancher/tests/v2/actions/rbac"
	//"github.com/rancher/rancher/tests/v2/actions/projects"
	rbacv2 "github.com/rancher/rancher/tests/v2/actions/kubeapi/rbac"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	extensionscluster "github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/clusters/kubernetesversions"
	//"github.com/rancher/shepherd/extensions/settings"
	"github.com/rancher/shepherd/extensions/users"
	"github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	kubeconfigSetting = "kubeconfig-default-token-ttl-minutes"
	updateValue       = "3"
	replacementGlobalRoleName = "restricted-admin-replacement"
)

type RestrictedAdminReplacementTestSuite struct {
	suite.Suite
	client  *rancher.Client
	session *session.Session
	cluster *management.Cluster
}

func (ra *RestrictedAdminReplacementTestSuite) TearDownSuite() {
	ra.session.Cleanup()
}

func (ra *RestrictedAdminReplacementTestSuite) SetupSuite() {
	ra.session = session.NewSession()

	client, err := rancher.NewClient("", ra.session)
	require.NoError(ra.T(), err)

	ra.client = client

	log.Info("Getting cluster name from the config file and append cluster details in the struct.")
	clusterName := client.RancherConfig.ClusterName
	require.NotEmptyf(ra.T(), clusterName, "Cluster name to install should be set")
	clusterID, err := extensionscluster.GetClusterIDByName(ra.client, clusterName)
	require.NoError(ra.T(), err, "Error getting cluster ID")
	ra.cluster, err = ra.client.Management.Cluster.ByID(clusterID)
	require.NoError(ra.T(), err)
}

var (
	replacementGlobalRole = v3.GlobalRole{
		ObjectMeta: v1.ObjectMeta{
			Name: replacementGlobalRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups: []string{"catalog.cattle.io"},
				Resources: []string{"clusterrepos"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"clustertemplates"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"clustertemplaterevisions"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"globalrolebindings"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"globalroles"},
				Verbs: []string{
					"delete", "deletecollection", "get", "list",
					"patch", "create", "update", "watch",
				},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"users", "userattribute", "groups", "groupmembers"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"podsecurityadmissionconfigurationtemplates"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"authconfigs"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"nodedrivers"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"kontainerdrivers"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"roletemplates"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"templates", "templateversions"},
				Verbs:     []string{"*"},
			},
		},
		InheritedClusterRoles: []string{
			"cluster-owner",
		},
		InheritedFleetWorkspacePermissions: &v3.FleetWorkspacePermission{
			ResourceRules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{"fleet.cattle.io"},
					Resources: []string{
						"clusterregistrationtokens", "gitreporestrictions", "clusterregistrations",
						"clusters", "gitrepos", "bundles", "bundledeployments", "clustergroups",
					},
					Verbs: []string{"*"},
				},
			},
			WorkspaceVerbs: []string{"get", "list", "update", "create", "delete"},
		},
	}
)

func createCustomGlobalRole(client *rancher.Client) (*v3.GlobalRole, error) {
	replacementGlobalRole.Name = namegen.AppendRandomString("testgr")
	createdGlobalRole, err := client.WranglerContext.Mgmt.GlobalRole().Create(&replacementGlobalRole)
	if err != nil {
		return nil, err
	}

	createdGlobalRole, err = rbac.GetGlobalRoleByName(client, createdGlobalRole.Name)
	if err != nil {
		return nil, err
	}

	return createdGlobalRole, err
}

func createCustomGlobalRoleAndUser(client *rancher.Client) (*v3.GlobalRole, *management.User, error) {
	createdGlobalRole, err := createCustomGlobalRole(client)

	createdUser, err := users.CreateUserWithRole(client, users.UserConfig(), rbac.StandardUser.String(), createdGlobalRole.Name)
	if err != nil {
		return nil, nil, err
	}

	return createdGlobalRole, createdUser, err
}


func updateGlobalSetting(client *rancher.Client, settingName string, settingValue string) error {
	setting, err := client.WranglerContext.Mgmt.Setting().Get(settingName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get setting %s: %w", settingName, err)
	}

	setting.Value = settingValue
	updatedSetting, err := client.WranglerContext.Mgmt.Setting().Update(setting)
	if err != nil {
		return fmt.Errorf("failed to update setting %s: %w", updatedSetting.Name, err)
	}
	return nil
}

func getGlobalSetting(client *rancher.Client, settingName string) (*v3.Setting, error) {
	setting, err := client.WranglerContext.Mgmt.Setting().Get(settingName, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return setting, nil
}

func (ra *RestrictedAdminReplacementTestSuite) TestRestrictedAdminReplacementCreateCluster() {
	subSession := ra.session.NewSession()
	defer subSession.Cleanup()

	log.Info("Create the replacement restricted admin global role")
	createdRAReplacementRole, createdUser, err := createCustomGlobalRoleAndUser(ra.client)
	require.NoError(ra.T(), err, "failed to create global role and user")

	createdRAReplacementUserClient, err := ra.client.AsUser(createdUser)
	require.NoError(ra.T(), err)

	
	// createdRAReplacementRole, err := rbacv2.CreateGlobalRole(ra.client, replacementGlobalRole)
	// require.NoError(ra.T(), err)
	
	// log.Info("Create a standard user and add the restricted admin replacement global role to the user")
	// _, restrictedAdminClient, err := rbac.SetupUser(ra.client, createdRAReplacementRole.Name)
	// require.NoError(ra.T(), err)
	// log.Info(restrictedAdminClient)


	ra.T().Logf("Validating user with %s role can create a downstream cluster", createdRAReplacementRole.Name)
	userConfig := new(provisioninginput.Config)
	config.LoadConfig(provisioninginput.ConfigurationFileKey, userConfig)
	nodeProviders := userConfig.NodeProviders[0]
	nodeAndRoles := []provisioninginput.NodePools{
		provisioninginput.AllRolesNodePool,
	}
	externalNodeProvider := provisioning.ExternalNodeProviderSetup(nodeProviders)
	clusterConfig := clusters.ConvertConfigToClusterConfig(userConfig)
	clusterConfig.NodePools = nodeAndRoles
	kubernetesVersion, err := kubernetesversions.Default(createdRAReplacementUserClient, extensionscluster.RKE1ClusterType.String(), []string{})
	require.NoError(ra.T(), err)

	clusterConfig.KubernetesVersion = kubernetesVersion[0]
	clusterConfig.CNI = userConfig.CNIs[0]
	clusterObject, _, err := provisioning.CreateProvisioningRKE1CustomCluster(createdRAReplacementUserClient, &externalNodeProvider, clusterConfig)
	require.NoError(ra.T(), err)
	provisioning.VerifyRKE1Cluster(ra.T(), createdRAReplacementUserClient, clusterConfig, clusterObject)
}

func (ra *RestrictedAdminReplacementTestSuite) TestRestrictedAdminReplacementGlobalSettings() {
	subSession := ra.session.NewSession()
	defer subSession.Cleanup()

	log.Info("Create the replacement restricted admin global role")
	replacementGlobalRoleName := "restricted-admin-replacement"
	replacementGlobalRole := &v3.GlobalRole{
		ObjectMeta: v1.ObjectMeta{
			Name: replacementGlobalRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups: []string{"catalog.cattle.io"},
				Resources: []string{"clusterrepos"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"clustertemplates"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"clustertemplaterevisions"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"globalrolebindings"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"globalroles"},
				Verbs: []string{
					"delete", "deletecollection", "get", "list",
					"patch", "create", "update", "watch",
				},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"users", "userattribute", "groups", "groupmembers"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"podsecurityadmissionconfigurationtemplates"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"authconfigs"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"nodedrivers"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"kontainerdrivers"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"roletemplates"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"templates", "templateversions"},
				Verbs:     []string{"*"},
			},
		},
		InheritedClusterRoles: []string{
			"cluster-owner",
		},
		InheritedFleetWorkspacePermissions: &v3.FleetWorkspacePermission{
			ResourceRules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{"fleet.cattle.io"},
					Resources: []string{
						"clusterregistrationtokens", "gitreporestrictions", "clusterregistrations",
						"clusters", "gitrepos", "bundles", "bundledeployments", "clustergroups",
					},
					Verbs: []string{"*"},
				},
			},
			WorkspaceVerbs: []string{"get", "list", "update", "create", "delete"},
		},
	}
	createdRAReplacementRole, err := rbacv2.CreateGlobalRole(ra.client, replacementGlobalRole)
	require.NoError(ra.T(), err)

	log.Info("Create a standard user and add the restricted admin replacement global role to the user")
	raReplacementUser, raReplacementUserClient, err := rbac.AddUserWithGlobalRole(ra.client, createdRAReplacementRole.Name)
	require.NoError(ra.T(), err)

	raReplacementUserContext, err := raReplacementUserClient.WranglerContext.DownStreamClusterWranglerContext(ra.cluster.ID)
	require.NoError(ra.T(), err)
	globalSettingsList, err := raReplacementUserContext.Mgmt.Setting().List(v1.ListOptions{})
	log.Infof("%s SETTINGS LIST: %+v", raReplacementUser.Name, globalSettingsList)

	adminUserContext, err := ra.client.WranglerContext.DownStreamClusterWranglerContext(ra.cluster.ID)
	adminGlobalSettingsList, err := adminUserContext.Mgmt.Setting().List(v1.ListOptions{})

	require.Equal(ra.T(), adminGlobalSettingsList, globalSettingsList)
	//require.Equal(ra.T(), len(adminGlobalSettingsList.Items), len(globalSettingsList.Items))
}

func (ra *RestrictedAdminReplacementTestSuite) TestRestrictedAdminReplacementCantUpdateGlobalSettings() {
	subSession := ra.session.NewSession()
	defer subSession.Cleanup()

	log.Info("Create the replacement restricted admin global role")
	replacementGlobalRoleName := "restricted-admin-replacement"
	replacementGlobalRole := &v3.GlobalRole{
		ObjectMeta: v1.ObjectMeta{
			Name: replacementGlobalRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups: []string{"catalog.cattle.io"},
				Resources: []string{"clusterrepos"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"clustertemplates"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"clustertemplaterevisions"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"globalrolebindings"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"globalroles"},
				Verbs: []string{
					"delete", "deletecollection", "get", "list",
					"patch", "create", "update", "watch",
				},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"users", "userattribute", "groups", "groupmembers"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"podsecurityadmissionconfigurationtemplates"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"authconfigs"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"nodedrivers"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"kontainerdrivers"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"roletemplates"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"management.cattle.io"},
				Resources: []string{"templates", "templateversions"},
				Verbs:     []string{"*"},
			},
		},
		InheritedClusterRoles: []string{
			"cluster-owner",
		},
		InheritedFleetWorkspacePermissions: &v3.FleetWorkspacePermission{
			ResourceRules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{"fleet.cattle.io"},
					Resources: []string{
						"clusterregistrationtokens", "gitreporestrictions", "clusterregistrations",
						"clusters", "gitrepos", "bundles", "bundledeployments", "clustergroups",
					},
					Verbs: []string{"*"},
				},
			},
			WorkspaceVerbs: []string{"get", "list", "update", "create", "delete"},
		},
	}
	createdRAReplacementRole, err := rbacv2.CreateGlobalRole(ra.client, replacementGlobalRole)
	require.NoError(ra.T(), err)

	log.Info("Create a standard user and add the restricted admin replacement global role to the user")
	log.Info("Create a project and a namespace in the project")
	//adminProject, _, err := projects.CreateProjectAndNamespace(ra.client, ra.cluster.ID)
	//require.NoError(ra.T(), err)
	//raReplacementUser, raReplacementUserClient, err := rbac.AddUserWithRoleToCluster(ra.client, createdRAReplacementRole.Name, rbac.StandardUser.String(), ra.cluster, adminProject)
	_, raReplacementUserClient, err := rbac.AddUserWithGlobalRole(ra.client, createdRAReplacementRole.Name)
	require.NoError(ra.T(), err)

	raReplacementUserContext, err := raReplacementUserClient.WranglerContext.DownStreamClusterWranglerContext(ra.cluster.ID)
	require.NoError(ra.T(), err)

	globalSettingToUpdate, err := raReplacementUserContext.Mgmt.Setting().Get(kubeconfigSetting, v1.GetOptions{})
	//log.Infof("%s globalSettingToUpdate: %+v", raReplacementUser.Name, globalSettingToUpdate)

	

	// updatedGlobalSetting := globalSettingToUpdate.DeepCopy()
	// updatedGlobalSetting.Value = "5"

	// updateGlobalSetting, err := raReplacementUserContext.Mgmt.Setting().Update(updatedGlobalSetting)
	// require.NoError(ra.T(), err)
	// log.Info(updateGlobalSetting)

	globalSettingToUpdate2, err := getGlobalSetting(ra.client, kubeconfigSetting)
	log.Info("GLOBAL SETTING TO UPDATE 2: ", globalSettingToUpdate2)
	require.NoError(ra.T(), err)

	
	ra.T().Logf("Validating restrictedAdmin replacement cannot edit global settings")
	err = updateGlobalSetting(raReplacementUserClient, globalSettingToUpdate.Name, updateValue)
	require.NoError(ra.T(), err)

	//Question for problem above: am I using the raReplacementUserContext when trying to update? 
}

// func (ra *RestrictedAdminReplacementTestSuite) TestRestrictedAdminReplacementGlobalSettings2() {

// 	subSession := ra.session.NewSession()
// 	defer subSession.Cleanup()

// 	log.Info("Create a global role with inheritedClusterRoles.")
// 	raReplacementGlobalRole := []string{rbac.ClusterOwner.String()}
// 	createdGlobalRole, err := createGlobalRoleWithInheritedClusterRolesWrangler(ra.client, raReplacementGlobalRole)
// 	require.NoError(ra.T(), err)

// 	log.Info("Create a user with global role standard user and custom global role.")
// 	raReplacementUser, err := users.CreateUserWithRole(ra.client, users.UserConfig(), rbac.StandardUser.String(), raReplacementGlobalRole.Name)
// 	require.NoError(ra.T(), err)

	
// 	ra.T().Log("Validating restricted Admin replacement can list global settings")
// 	steveRestrictedAdminclient := restrictedAdminClient.Steve
// 	steveAdminClient := ra.client.Steve

// 	adminListSettings, err := steveAdminClient.SteveType(settings.ManagementSetting).List(nil)
// 	require.NoError(ra.T(), err)
// 	adminSettings := adminListSettings.Names()

// 	resAdminListSettings, err := steveRestrictedAdminclient.SteveType(settings.ManagementSetting).List(nil)
// 	require.NoError(ra.T(), err)
// 	resAdminSettings := resAdminListSettings.Names()

// 	assert.Equal(ra.T(), len(adminSettings), len(resAdminSettings))
// 	assert.Equal(ra.T(), adminSettings, resAdminSettings)
// }

func TestRestrictedAdminReplacementTestSuite(t *testing.T) {
	suite.Run(t, new(RestrictedAdminReplacementTestSuite))
}
