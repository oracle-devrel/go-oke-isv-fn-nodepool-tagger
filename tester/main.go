//  ISV Serverless Nodepool Tagger, Ed Shnekendorf, September 2021
//  Copyright (c) 2021 Oracle and/or its affiliates.
//  Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/common/auth"
	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/oracle/oci-go-sdk/objectstorage"
)

func main() {
	// create SDK provider with ResourcePrincipal permissions
	var provider common.ConfigurationProvider
	provider, err := auth.InstancePrincipalConfigurationProvider()
	helpers.FatalIfError(err)
	//provider = common.DefaultConfigProvider()

	// generate file that will be written to OS
	file, err := ioutil.TempFile("/tmp", "os")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
	_, err = file.WriteString("a small file targeted for object storage\n")
	if err != nil {
		log.Fatal(err)
	}
	file.Sync()
	info, err := file.Stat()

	// create OS service client
	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)

	// get OS namespace
	req := objectstorage.GetNamespaceRequest{}
	resp, err := client.GetNamespace(context.Background(), req)
	helpers.FatalIfError(err)

	// attempt to write to bucket-a
	bucketName := "bucket-a"
	file, err = os.Open(file.Name())
	defer file.Close()
	putReq := objectstorage.PutObjectRequest{ContentLength: common.Int64(info.Size()), PutObjectBody: file,
		NamespaceName: common.String(*resp.Value), BucketName: common.String(bucketName),
		ObjectName: common.String("testfile")}
	_, err = client.PutObject(context.Background(), putReq)

	if err == nil {
		fmt.Printf("File written to storage bucket: " + bucketName + "\n")
	} else {
		fmt.Printf("Failed to write file to bucket: " + err.Error() + "\n")
	}

	// attempt to write to bucket-b
	bucketName = "bucket-b"
	file, err = os.Open(file.Name())
	defer file.Close()
	putReq = objectstorage.PutObjectRequest{ContentLength: common.Int64(info.Size()), PutObjectBody: file,
		NamespaceName: common.String(*resp.Value), BucketName: common.String(bucketName),
		ObjectName: common.String("testfile")}
	_, err = client.PutObject(context.Background(), putReq)

	if err == nil {
		fmt.Printf("File written to storage bucket: " + bucketName + "\n")
	} else {
		fmt.Printf("Failed to write file to bucket: " + err.Error() + "\n")
	}

	// loop infinitely to keep container alive
	for {
	}
}
