# Automating a Pod Identity Solution with Oracle Container Engine for Kubernetes (OKE)

[![License: UPL](https://img.shields.io/badge/license-UPL-green)](https://img.shields.io/badge/license-UPL-green) [![Quality gate](https://sonarcloud.io/api/project_badges/quality_gate?project=oracle-devrel_go-oke-isv-fn-nodepool-tagger)](https://sonarcloud.io/dashboard?id=oracle-devrel_go-oke-isv-fn-nodepool-tagger)

## Introduction
Oracle Container Engine for Kubernetes is a great service to host your microservices as it provides a certified Kubernetes environment and a full range of native Oracle Cloud Infrastructure services that can be leveraged by your custom stack.  As the complexity of your application grows, you may want to consider introducing fine-grained security for your pods to OCI services.

Let's suppose that you're developing containerized applications and hosting them on Oracle's Container Engine for Kubernetes (OKE).  Let's also suppose that your custom code, running inside a Kubernetes Pod, needs fine-grained access to Oracle Cloud Infrastructure (OCI) services like streams, object storage, or cloud functions. 

For example, some portion of your application may handle accounts receivables data while another portion handles accounts payables and for compliance purposes you have segregated this data into two separate object storage buckets with different policy-based permissions.  How do you control access so that pods hosting payables microservices can only access the payables bucket while pods hosting receivables microservices may only access the appropriate bucket? 

One approach would be to create a service account in OCI Identity & Access Management (IAM) and inject those credentials into your microservices using a Kubernetes Secret so that your code can use them to authorize via the OCI SDK.  Note that Kubernetes Secrets in OKE can be made even more secure by encrypting them at rest using OCI Vault.  However, many organizations frown upon the use of service accounts and utilizing this approach still requires you to manage credentials as part of your CI/CD process, handle them securely in your code, and implement password rotation and revocation.

A more sophisticated approach leverages the power of OCI Instance Principals to remove the need to create service accounts or pass credentials, instead tying authorization to a dynamic group based on tags applied to the worker nodes within distinct OKE node pools and the matching of pods with the appropriate node pools by leveraging node selector capabilities.  This approach is described in the OKE documentation under the Controlling Which Nodes Pods May Access and Limit Access Given to Instance Principals sections.

## Getting Started
### Create tag namespace and tag definition to classify node pools.  

In this sample namespace is 'oke-nodepool-tags' and defined tag key name is 'oke-nodepool-name'. Replace mycompartment with your compartment OCID in the first command use the output of that command to populate mytagnamespaceid in the second.

`oci iam tag-namespace create --compartment-id "ocid1.compartment.oc1..mycompartment" --name "oke-nodepool-tags" --description "Namespace for OKE nodepool instances"`

`oci iam tag create --tag-namespace-id "ocid1.tagnamespace.oc1..mytagnamespaceid" --name "oke-nodepool-name" --description "Name of the OKE nodepool this instance is associated with"`

### Create dynamic groups for each node pool.  
In this sample using names "pool-a" and "pool-b" which correspond to the nodepool names set in OKE.  The nodepool name also automatically applies a Kubernetes label to the node.  Make sure to replace compartment OCIDs.

`oci iam dynamic-group create --name "oke-pool-a-dg" --description "All worker nodes in OKE pool-a" --matching-rule "All {instance.compartment.id='ocid1.compartment.oc1..mycompartment',tag.oke-nodepool-tags.oke-nodepool-name.value='pool-a'}"`

`oci iam dynamic-group create --name "oke-pool-b-dg" --description "All worker nodes in OKE pool-b" --matching-rule "All {instance.compartment.id='ocid1.compartment.oc1..mycompartment',tag.oke-nodepool-tags.oke-nodepool-name.value='pool-b'}"`


### Create dynamic group for function and policy granting function required OCI permissions. 

Make sure to replace compartment name (xxxx) and compartment OCID (mycompartment)

`oci iam dynamic-group create --name "all-functions-in-xxxx-compartment" --description "All functions in compartment xxxx" --matching-rule "All {resource.type = 'fnfunc', resource.compartment.id='ocid1.compartment.oc1..mycompartment'}"`


### Create policies allowing function appropriate rights to read data from tenancy and update tags on compute instances

These are the policy statements created by the command.  Make sure to replace compartment name (xxxx) and compartment OCID (mycompartment)
- Allow dynamic-group all-functions-in-xxxx-compartment to use tag-namespaces in compartment xxxx	
- Allow dynamic-group all-functions-in-xxxx-compartment to read all-resources in compartment xxxx	
- Allow dynamic-group all-functions-in-xxxx-compartment to manage instance-family in compartment xxxx	where request.permission = 'INSTANCE_UPDATE'

`oci iam policy create --compartment-id "ocid1.compartment.oc1..mycompartment" --name "all-functions-in-xxxx-policy" --description "Permissions for function resources in xxxx compartment" --statements '["Allow dynamic-group all-functions-in-xxxx-compartment to use tag-namespaces in compartment xxxx","Allow dynamic-group all-functions-in-xxxx-compartment to read all-resources in compartment xxxx", "Allow dynamic-group all-functions-in-xxxx-compartment to manage instance-family in compartment xxxx"]'`


### Create the function, set the 3 required configuration parameters, and enable function logging

Although this can be automated easiest way is to follow the quickstart to create a functions application, enable logging on the application, and then follow the "Getting Started" section in the console under application to create and push a function.  Before pushing simply replace the contents of your function directory with this project

Create an event that invokes the function for every compute instance that is created in the compartment.  Make sure to replace compartment OCID (mycompartment) and function OCID (myfunction)

`oci events rule create --compartment-id "ocid1.compartment.oc1..mycompartment" --display-name "isv-nodepool-tagger" --description "On detect of compute instance creation check to see if it is part of an OKE nodepool and if a match exists tag the instance with the name of the nodepool" --is-enabled true --condition "{\"eventType\": [\"com.oraclecloud.computeapi.launchinstance.end\"]}" --actions "{\"actions\": [{\"action-type\": \"FAAS\",\"description\": null,\"function-id\": \"ocid1.fnfunc.oc1.iad.myfunction\",\"is-enabled\": true}]}"`

### Create object storage buckets and policies for testing

Create two buckets (bucket-a and bucket-b) and write policy that enables oke-pool-a-dg to interact only with bucket-a and 
oke-pool-b-dg to interact only with bucket-b.  Make sure to replace compartment OCID (mycompartment) and compartment name (xxxx)

`oci os bucket create --compartment-id "ocid1.compartment.oc1..mycompartment" --name "bucket-a"`

`oci os bucket create --compartment-id "ocid1.compartment.oc1..mycompartment" --name "bucket-b"`

`oci iam policy create --compartment-id "ocid1.compartment.oc1..mycompartment" --name "oke-nodepool-bucket-test-policy" --description "oke-pool-dg-a -> bucket-a and oke-pool-dg-b -> bucket-b" --statements '["Allow dynamic-group oke-pool-a-dg to manage objects in compartment xxxx where all {target.bucket.name='\''bucket-a'\''}","Allow dynamic-group oke-pool-b-dg to manage objects in compartment xxxx where all {target.bucket.name='\''bucket-b'\''}"]'`


### Test Steps on Compute

You can first test that *InstancePrincipal* authentication works from the worker nodes themselves by SSHing into compute instances and using the OCI CLI to attempt to manipulate objects inside a bucket.  

1) ssh to the OKE node in pool a
2) `touch testfile`
3) `oci os object put --auth instance_principal --bucket-name "bucket-a" --file ./testfile --force`
4) see a sucess message
5) `oci os object put --auth instance_principal --bucket-name "bucket-b" --file ./testfile --force`
6) see an authorization failure
7) ssh to the OKE node in pool b and run the same test with the expectation of inverted results

### Test steps with OKE

Once you've verified at the compute level, you can test that k8s pods running on particular nodes inherit the same *InstancePrincipal* permissions of their nodes.  

Execute the following steps while making sure to replace the **myregion**, **mytenancyname**, **my.name@email.com**, and **myauthtoken** keys with appropriate values for your environment in the statements below *and* in the deploy-to-oke.yaml file.

1. `cd tester`

2. `kubectl create secret docker-registry ocirsecret \
    --docker-server=myregion.ocir.io  \
    --docker-username='mytenancyname/oracleidentitycloudservice/my.name@email.com' \
    --docker-password='myauthtoken'  --docker-email='my.name@email.com'`

3. `docker build --tag myregion.ocir.io/mytenancyname/isv/docker-os-tester .`

4. `docker push myregion.ocir.io/mytenancyname/isv/docker-os-tester`

5. `kubectl apply -f deploy-to-oke.yaml`

### CLI Reference

#### Get list of nodepools

##### Command:

`oci ce node-pool list --compartment-id "ocid1.compartment.oc1..mycompartment" --query "data [*].{poolname:\"name\", ocid:id}"`

##### Sample Result:

```
[
  {
    "ocid": "ocid1.nodepool.oc1.iad.aaaa",
    "poolname": "pool-a"
  },
  {
    "ocid": "ocid1.nodepool.oc1.iad.bbbb",
    "poolname": "pool-b"
  }
]
```

#### Get all nodes from a pool; do this for each of the pools returned from previous command

##### Command:

`oci ce node-pool get --node-pool-id "ocid1.nodepool.oc1.iad.mynodepool" --query "data.nodes [*].{privateip:\"private-ip\", ocid:id}"`

##### Sample Result: 

```
[
  {
    "ocid": "ocid1.instance.oc1.iad.n1n1n1",
    "privateip": "10.0.10.243"
  },
  {
    "ocid": "ocid1.instance.oc1.iad.n2n2n2",
    "privateip": "10.0.10.163"
  }
]
```

#### Once you find a match from looping through the previous results, add correct tag to instance

##### Command:

`oci compute instance update --force --instance-id "ocid1.instance.oc1.iad.n1n1n1" --defined-tags "{\"Oracle-Tags\": {\"CreatedBy\": \"oke\"},\"oke-nodepool-tags\": {\"oke-nodepool-name\": \"pool-a\"}}"`

## Prerequisites
You have already provisioned an OKE cluster to use for your application and this sample demonstrates the concepts using a cluster with two node pools named pool-a and pool-b; make sure as a starting point both of these pools have 0 nodes in them.  You also have a working understanding of Oracle Cloud Functions as well as experience working with OCI identity constructs like policies and dynamic groups.

You will also need to have the latest version of the OCI Command Line Interface (CLI) installed and configured or use the auto-configured version provided for you by OCI Cloud Shell.  Keep in mind that you can customize all names (node pool names, tag names, dynamic group names, function names, etc) when you integrate this into your own environment.  

## Notes/Issues
* Nothing at this time

## URLs
* Nothing at this time

## Contributing
This project is open source.  Please submit your contributions by forking this repository and submitting a pull request!  Oracle appreciates any contributions that are made by the open source community.

## License
Copyright (c) 2024 Oracle and/or its affiliates.

Licensed under the Universal Permissive License (UPL), Version 1.0.

See [LICENSE](LICENSE.txt) for more details.
