package lastusedat

import (
	"time"
	"testing"
	"github.com/rancher/rancher/tests/v2/actions/kubeapi/tokens"
	"github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/kubeconfig"
	"github.com/rancher/shepherd/clients/rancher"
	users "github.com/rancher/shepherd/extensions/users"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

)

const (
	catAPIEndPoint = "cluster.cattle.io.clusterauthtoken"
	StandardUser = "user"
	ClusterOwner = "cluster-owner"
	LocalCluster = "local"
	EmptyString  = ""
	cattleSystem = "cattle-system"
)

type LastUsedAtTestSuite struct {
	suite.Suite
	client  *rancher.Client
	session *session.Session
	cluster *management.Cluster
}

func (lua *LastUsedAtTestSuite) TearDownSuite() {
	lua.session.Cleanup()
}

func (lua *LastUsedAtTestSuite) SetupSuite() {
	testSession := session.NewSession()
	lua.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(lua.T(), err)

	lua.client = client

	log.Info("Getting cluster name from the config file and append cluster details in lastusedat")
	clusterName := client.RancherConfig.ClusterName
	require.NotEmptyf(lua.T(), clusterName, "Cluster name to install should be set")
	clusterID, err := clusters.GetClusterIDByName(lua.client, clusterName)
	require.NoError(lua.T(), err, "Error getting cluster ID")
	lua.cluster, err = lua.client.Management.Cluster.ByID(clusterID)
	require.NoError(lua.T(), err)

}

func (lua *LastUsedAtTestSuite) TestLastUsedAtTime() {
	log.Info("Create a standard user and login to the cluster.")
	standardUser, err := users.CreateUserWithRole(lua.client, users.UserConfig(), "user")
	require.NoError(lua.T(), err, "Failed to create standard user")
	log.Info("User created: ", standardUser.Name)
	standardUserClient, err := lua.client.AsUser(standardUser)
	require.NoError(lua.T(), err)
	_, err = standardUserClient.ReLogin()
	require.NoError(lua.T(), err)

 	userToken, err := getTokenByUser(lua.client, *standardUser)
	require.NoError(lua.T(), err)

	log.Info("Verify LastUsedAt field is updated in the token object")
	require.False(lua.T(), userToken.LastUsedAt.IsZero(), "LastUsedAt is empty")
	lua.T().Logf("Token: %v was last used at: %s", userToken.Name, userToken.LastUsedAt.Time)
}

func (lua *LastUsedAtTestSuite) TestLastUsedAtACEEnabledFullContext() {
	log.Info("Create a standard user and login to the cluster.")
	standardUser, err := users.CreateUserWithRole(lua.client, users.UserConfig(), "user")
	require.NoError(lua.T(), err, "Failed to create standard user")
	log.Info("User created: ", standardUser.Name)
	
	log.Info("Login as the standard user")
	standardUserClient, err := lua.client.AsUser(standardUser)
	require.NoError(lua.T(), err)
	_, err = standardUserClient.ReLogin()
	require.NoError(lua.T(), err)

	downstreamContext, err := lua.client.WranglerContext.DownStreamClusterWranglerContext(lua.cluster.ID)
	require.NoError(lua.T(), err)
	clusterAuthTokenList, err := downstreamContext.Cluster.ClusterAuthToken().List(namespace, v1.ListOptions{})
	require.NoError(lua.T(), err)
	log.Info("ClusterAuthTokenList =>", clusterAuthTokenList)

}

func (lua *LastUsedAtTestSuite) TestLastUsedAtACEEnabledNormalContext() {
	log.Info("Create an admin user and login to the cluster.")
	clusterOwner, err := users.CreateUserWithRole(lua.client, users.UserConfig(), "admin")
	require.NoError(lua.T(), err, "Failed to create admin user")
	log.Info("User created: ", clusterOwner.Name)
	clusterOwnerClient, err := lua.client.AsUser(clusterOwner)
	require.NoError(lua.T(), err)
	_, err = clusterOwnerClient.ReLogin()
	require.NoError(lua.T(), err)

	generateUserKubeConfig, err := kubeconfig.GetKubeconfig(clusterOwnerClient, lua.cluster.ID)
	require.NoError(lua.T(), err)
	lua.T().Logf("Kubeconfig %v generated for user: %v", generateUserKubeConfig, clusterOwner)


	steveClient, err := lua.client.Steve.ProxyDownstream(lua.cluster.ID)
	assert.NoError(lua.T(), err)
	authTokenList, err := steveClient.SteveType(catAPIEndPoint).List(nil)
	require.NoError(lua.T(), err)

	userClusterAuthTokens := []string{}
	for _, authToken := range authTokenList.Data {
		if authToken.JSONResp["userName"] == clusterOwner.ID {
			userClusterAuthTokens = append(userClusterAuthTokens, authToken.Name)
		}
	}
	log.Info("User's ClusterAuthTokens", userClusterAuthTokens)

	log.Info("Verify that ClusterAuthToken doesn't contain LastUsedAt field since normal context is used")
	assert.NotContains(lua.T(), (&userClusterAuthTokens), LastUsedAtLabel)

 	// userToken, err := getClusterAuthTokenByUser(lua.client, *clusterOwner)
	// require.NoError(lua.T(), err)


	//log.Info("Verify  LastUsedAt field is updated in the token object")
	// require.True(lua.T(), userToken.LastUsedAt.IsZero(), "LastUsedAt is empty")
	// // log.Info("Verify LastUsedAt field is doesn't exist in the ClusterAuthToken object")
	// // require.False(lua.T(), userToken.LastUsedAt.IsZero(), "LastUsedAt is empty")
	// lua.T().Logf("Token: %v was last used at: %s", userToken.Name, userToken.LastUsedAt.Time)
}

func (lua *LastUsedAtTestSuite) TestLastUsedAtTimeWebhookChecks() {
	log.Info("Create a standard user and login to the cluster.")
	standardUser, err := users.CreateUserWithRole(lua.client, users.UserConfig(), "user")
	require.NoError(lua.T(), err, "Failed to create standard user")
	log.Info("User created: ", standardUser.Name)
	standardUserClient, err := lua.client.AsUser(standardUser)
	require.NoError(lua.T(), err) 
	_, err = standardUserClient.ReLogin()
	require.NoError(lua.T(), err)

 	userToken, err := getTokenByUser(lua.client, *standardUser)
	require.NoError(lua.T(), err)

	updatedTime := time.Now()
	updatedTimeString := updatedTime.Format(time.RFC3339)
	invalidTimeString := updatedTime.Format(time.RFC1123)

	log.Info("Update LastUsedAt field in token object to a valid RFC3339 value : ", updatedTimeString)
	patchedTokenWithValidLastUsedTime, unstructuredResValidLastUsedTime, err := tokens.PatchToken(lua.client, LocalCluster, userToken.Name, "replace", "/lastUsedAt", updatedTimeString)
	assert.NotEqual(lua.T(), patchedTokenWithValidLastUsedTime.LastUsedAt.String(), userToken.LastUsedAt.String())
	require.NoError(lua.T(), err)

	log.Info("Update LastUsedAt field in token object to an invalid RFC1123 value : ", invalidTimeString)
	patchedTokenWithInvalidLastUsedAtTime, unstructuredResInvalidLastUsedTime, err := tokens.PatchToken(lua.client, LocalCluster, userToken.Name, "replace", "/lastUsedAt", invalidTimeString)
	require.Error(lua.T(), err)

	log.Info("Update LastUsedAt field in token object to an invalid value (empty string)")
	patchedTokenWithInvalidLastUsedAtValue, unstructuredResInvalidLastUsedValue, err := tokens.PatchToken(lua.client, LocalCluster, userToken.Name, "replace", "/lastUsedAt", EmptyString)
	require.Error(lua.T(), err)

	log.Info("Patched token->", patchedTokenWithValidLastUsedTime, unstructuredResValidLastUsedTime)
	log.Info("Patched token->", patchedTokenWithInvalidLastUsedAtTime, unstructuredResInvalidLastUsedTime)
	log.Info("Patched token->",patchedTokenWithInvalidLastUsedAtValue, unstructuredResInvalidLastUsedValue)

}

func TestLastUsedAtTestSuite(t *testing.T) {
	suite.Run(t, new(LastUsedAtTestSuite))
}
