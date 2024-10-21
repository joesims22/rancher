package lastusedat

import (
	"fmt"
	// "time"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/shepherd/clients/rancher"
	//"github.com/rancher/rancher/tests/v2/actions/kubeapi/tokens"
	client "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//log "github.com/sirupsen/logrus"
	//clusterv3 "github.com/rancher/rancher/pkg/apis/cluster.cattle.io/v3"
)

const (
	LastUsedAtLabel = "/lastUsedAt"
	namespace    = "cattle-system"
)

func getTokenByUser(rancherClient *rancher.Client, user client.User) (*v3.Token, error) {
	_, err := rancherClient.AsUser(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to set user context: %v", err)
	}

	tokenList, err := rancherClient.WranglerContext.Mgmt.Token().List(v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list tokens: %v", err)
	}

	// Filter tokens by user ID
	for _, token := range tokenList.Items {
		if token.Labels["authn.management.cattle.io/token-userId"] == user.ID {
			return &token, nil
		}
	}

	return nil, fmt.Errorf("no tokens found for user: %s", user.ID)
}

// func getClusterAuthTokenByUser(rancherClient *rancher.Client, user client.User) (*clusterv3.ClusterAuthToken, error) {
// 	_, err := rancherClient.AsUser(&user)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to set user context: %v", err)
// 	}
// 	// downstreamContext, err := rancherClient.WranglerContext.DownStreamClusterWranglerContext(cluster.ID)
// 	// require.NoError(err)
// 	// clusterAuthTokenList, err := downstreamContext.Cluster.ClusterAuthToken().List(namespace, v1.ListOptions{})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to list tokens: %v", err)
// 	}

// 	// Filter tokens by user ID
// 	for _, token := range clusterAuthTokenList.Items {
// 		if token.Labels["authn.management.cattle.io/token-userId"] == user.ID {
// 			return &token, nil
// 		}
// 	}

// 	return nil, fmt.Errorf("no tokens found for user: %s", user.ID)
// }



// func getCATByUser(rancherClient *rancher.Client, namespace string, user client.User) (*v3.ClusterRegistrationToken, error) {
// 	_, err := rancherClient.AsUser(&user)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to set user context: %v", err)
// 	}

// 	CatList, err := rancherClient.WranglerContext.
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to list tokens: %v", err)
// 	}

// 	// Filter tokens by user ID
// 	for _, token := range CatList.Items {
// 		if token.Labels["authn.management.cattle.io/token-userId"] == user.ID {
// 			return &token, nil
// 		}
// 	}

// 	return nil, fmt.Errorf("no tokens found for user: %s", user.ID)
// }


// getClusterAuthTokensByUserID retrieves ClusterAuthTokens for a given userID.
// func getClusterAuthTokensByUserID(rancherClient *rancher.Client, userID string, clusterName string) ([]*clusterv3.ClusterAuthToken, error) {
//     // Assuming you have a way to access the specific cluster context
//     context, err := rancherlient.WranglerContext.DownStreamClusterWranglerContext(rbs.cluster.ID)
//     assert.NoError(rbs.T(), err)
//     clusterAuthToken, err := context.Mgmt.ClusterRegistrationToken().List()
//     if err != nil {
//         return nil, fmt.Errorf("failed to get cluster context: %v", err)
//     }

//     // Use label selector to filter by userName (assuming userID maps to userName)
//     listOptions := v1.ListOptions{
//         LabelSelector: labels.Set{"userName": userID}.String(),
//     }

//     // List ClusterAuthTokens in the specified cluster
//     tokenList, err := clusterContext.Mgmt.ClusterAuthToken().List(listOptions)
//     if err != nil {
//         return nil, fmt.Errorf("failed to list cluster auth tokens: %v", err)
//     }

//     // If there are no tokens, return an empty slice
//     if len(tokenList.Items) == 0 {
//         return nil, nil
//     }

//     // Return the list of tokens
//     return tokenList.Items, nil
// }
