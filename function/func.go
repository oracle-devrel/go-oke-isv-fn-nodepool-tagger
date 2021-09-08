//	ISV Serverless Nodepool Tagger, Ed Shnekendorf, September 2021
//  Copyright (c) 2021 Oracle and/or its affiliates.
//  Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	fdk "github.com/fnproject/fdk-go"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/common/auth"
	"github.com/oracle/oci-go-sdk/containerengine"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/example/helpers"
)

func main() {
	fdk.Handle(fdk.HandlerFunc(myHandler))
}

//
// Main Fn entrypoint
//
func myHandler(ctx context.Context, in io.Reader, out io.Writer) {
	// decode cloud event
	var evt OCIEvent
	json.NewDecoder(in).Decode(&evt)
	fmt.Printf("Got OCI event with EventType [%s] and resourceID [%s]\n", evt.EventType, evt.Data.ResourceID)
	instanceId := evt.Data.ResourceID

	// working variables pulled from cloud event or config
	fnCtx := fdk.GetContext(ctx)

	// get config from function context
	tagNamespaceName := fnCtx.Config()["tag_namespace"]
	tagName := fnCtx.Config()["tag_name"]
	compartmentId := fnCtx.Config()["compartment_id"]
	match := false
	if len(tagNamespaceName) < 1 || len(tagName) < 1 || len(compartmentId) < 1 {
		helpers.FatalIfError(errors.New("tag_namespace, tag_name, and compartment_id must be defined in Function config"))
		return
	}

	// create SDK provider with ResourcePrincipal permissions
	var provider common.ConfigurationProvider
	provider, err := auth.ResourcePrincipalConfigurationProvider()
	helpers.FatalIfError(err)

	// create OKE service client
	client, err := containerengine.NewContainerEngineClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)

	// get all nodepools in the compartment
	req := containerengine.ListNodePoolsRequest{CompartmentId: common.String(compartmentId)}
	resp, err := client.ListNodePools(context.Background(), req)
	helpers.FatalIfError(err)

	// for each pool in the compartment...
	for _, pool := range resp.Items {
		req := containerengine.GetNodePoolRequest{NodePoolId: pool.Id}
		resp, err := client.GetNodePool(context.Background(), req)
		helpers.FatalIfError(err)

		// for each node in the pool
		for _, node := range resp.Nodes {
			// if an an instance in the pool is found matching the desired instance...
			if instanceId == *node.Id {
				fmt.Printf("MATCH: pool[%s] matches node id[%s]\n", *pool.Name, *node.Id)
				match = true

				// tag the instance with the associated OKE pool name.  also recreate the default 'created-by' tag
				client, err := core.NewComputeClientWithConfigurationProvider(provider)
				helpers.FatalIfError(err)
				req := core.UpdateInstanceRequest{
					InstanceId: node.Id,
					UpdateInstanceDetails: core.UpdateInstanceDetails{
						DefinedTags: map[string]map[string]interface{}{tagNamespaceName: {tagName: *pool.Name}, "Oracle-Tags": {"CreatedBy": "oke"}}}}
				_, err = client.UpdateInstance(context.Background(), req)
				helpers.FatalIfError(err)

				fmt.Printf("MATCH: node id[%s] has been tagged\n", *node.Id)

			}
		}
	}

	// log lack of match between instance and node pool
	if !match {
		fmt.Printf("NO-MATCH: No match for instance id[%s] found.  No nodes tagged.\n", instanceId)
	}

	return
}
