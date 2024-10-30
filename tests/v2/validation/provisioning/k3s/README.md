# K3S Provisioning Configs

For your config, you will need everything in the Prerequisites section on the previous readme, [Define your test](#provisioning-input), and at least one [Cloud Credential](#cloud-credentials) and [Node Driver Machine Config](#machine-k3s-config) or [Custom Cluster Template](#custom-cluster), which should match what you have specified in `provisioningInput`. 

Your GO test_package should be set to `provisioning/k3s`.
Your GO suite should be set to `-run ^TestK3SProvisioningTestSuite$`.
Please see below for more details for your config. Please see below for more details for your config. Please note that the config can be in either JSON or YAML (all examples are illustrated in YAML).

## Table of Contents
1. [Prerequisites](../README.md)
2. [Configuring test flags](#Flags)
3. [Define your test](#provisioning-input)
4. [Cloud Credential](#cloud-credentials)
5. [Configure providers to use for Node Driver Clusters](#machine-k3s-config)
6. [Configuring Custom Clusters](#custom-cluster)
7. [Template Test](#template-test)
8. [Static test cases](#static-test-cases)
9. [Advanced Cluster Settings](#advanced-settings)
10. [Back to general provisioning](../README.md)

## Flags
Flags are used to determine which static table tests are run (has no effect on dynamic tests) 
`Long` Will run the long version of the table tests (usually all of them)
`Short` Will run the subset of table tests with the short flag.

```yaml
flags:
  desiredflags: "Long"   #required (static tests only)
```

## Provisioning Input
provisioningInput is needed to the run the K3S tests.

**nodeProviders is only needed for custom cluster tests; the framework only supports custom clusters through aws/ec2 instances.**
```yaml
provisioningInput:
  machinePools:                                              
  - machinePoolConfig:                       #required(dynamic only) (at least 1 of each role must be true accross all machinePoolConfigs)
      etcd: true                             #required(dynamic only) (at least 1 role etcd, controlplane, worker must be true)
      controlplane: true
      worker: true
      quantity: 1
      drainBeforeDelete: true
      hostnameLengthLimit: 29
      nodeStartupTimeout: "600s"
      unhealthyNodeTimeout: "300s"
      maxUnhealthy: "2"
      unhealthyRange: "2-4"
  - machinePoolConfig:
      worker: true
      quantity: 2
      drainBeforeDelete: true
  - machinePoolConfig:
      worker: true
      quantity: 1
  k3sKubernetesVersion: ["v1.27.6+k3s1"]     #required (at least 1)
  providers: ["aws"]                         #required (at least 1) linode,aws,do,harvester,vsphere,azure,google
  cni: ["calico"]                            #required (at least 1)
  nodeProviders: ["ec2"]                     #required(custom clusters only)
  hardened: false
  psact: ""                                  #either rancher-privileged|rancher-restricted|rancher-baseline
  clusterSSHTests: ["CheckCPU", "NodeReboot", "AuditLog"]
  etcd:
    disableSnapshot: false
    snapshotScheduleCron: "0 */5 * * *"
    snapshotRetain: 3
    s3:
      bucket: ""
      endpoint: "s3.us-east-2.amazonaws.com"
      endpointCA: ""
      folder: ""
      region: "us-east-2"
      skipSSLVerify: true
```

## Cloud Credentials
These are the inputs needed for the different node provider cloud credentials, including linode, aws, harvester, azure, and google.

### Digital Ocean
```yaml
digitalOceanCredentials:               
  accessToken: ""                     #required
```
### Linode
```yaml
linodeCredentials:                   
  token: ""                           #required
```
### Azure
```yaml
azureCredentials:                     
  clientId: ""                        #required
  clientSecret: ""                    #required
  subscriptionId": ""                 #required
  environment: "AzurePublicCloud"     #required
```
### AWS
```yaml
awsCredentials:                       
  secretKey: ""                       #required
  accessKey: ""                       #required
  defaultRegion: ""                   #required
```
### Harvester
```yaml
harvesterCredentials:                 
  clusterId: ""                       #required
  clusterType: ""                     #required
  kubeconfigContent: ""               #required
```
### Google
```yaml
googleCredentials:                    
  authEncodedJson: ""                 #required
```
### VSphere
```yaml
vmwarevsphereCredentials:             
  password: ""                        #required
  username: ""                        #required
  vcenter: ""                         #required
  vcenterPort: ""                     #required
```

## Machine K3S Config
Machine K3S config is the final piece needed for the config to run K3S provisioning tests.

### AWS K3S Machine Config
```yaml
awsMachineConfigs:
  region: "us-east-2"                         #required
  awsMachineConfig:
  - roles: ["etcd","controlplane","worker"]   #required
    ami: ""                                   #required
    instanceType: "t3a.medium"                
    sshUser: "ubuntu"                         #required
    vpcId: ""                                 #required
    volumeType: "gp2"                         
    zone: "a"                                 #required
    retries: "5"                              
    rootSize: "60"                            
    securityGroup: [""] 
```
### Digital Ocean K3S Machine Config
```yaml
doMachineConfig:
  image: "ubuntu-20-04-x64"
  backups: false
  ipv6: false
  monitoring: false
  privateNetworking: false
  region: "nyc3"
  size: "s-2vcpu-4gb"
  sshKeyContents: ""
  sshKeyFingerprint: ""
  sshPort: "22"
  sshUser: "root"
  tags: ""
  userdata: ""
```
### Linode K3S Machine Config
```yaml
linodeMachineConfig:
  authorizedUsers: ""
  createPrivateIp: true
  dockerPort: "2376"
  image: "linode/ubuntu22.04"
  instanceType: "g6-standard-8"
  region: "us-west"
  rootPass: ""
  sshPort: "22"
  sshUser: ""
  stackscript: ""
  stackscriptData: ""
  swapSize: "512"
  tags: ""
  uaPrefix: "Rancher"
```
### Azure K3S Machine Config
```yaml
azureMachineConfig:
  availabilitySet: "docker-machine"
  diskSize: "30"
  environment: "AzurePublicCloud"
  faultDomainCount: "3"
  image: "canonical:UbuntuServer:22.04-LTS:latest"
  location: "westus"
  managedDisks: false
  noPublicIp: false
  nsg: ""
  openPort: ["6443/tcp", "2379/tcp", "2380/tcp", "8472/udp", "4789/udp", "9796/tcp", "10256/tcp", "10250/tcp", "10251/tcp", "10252/tcp"]
  resourceGroup: "docker-machine"
  size: "Standard_D2_v2"
  sshUser: "docker-user"
  staticPublicIp: false
  storageType: "Standard_LRS"
  subnet: "docker-machine"
  subnetPrefix: "192.168.0.0/16"
  updateDomainCount: "5"
  usePrivateIp: false
  vnet: "docker-machine-vnet"
```
### Harvester K3S Machine Config
```yaml
harvesterMachineConfig":
  diskSize: "40"
  cpuCount: "2"
  memorySize: "8"
  networkName: "default/ctw-network-1"
  imageName: "default/image-rpj98"
  vmNamespace: "default"
  sshUser: "ubuntu"
  diskBus: "virtio
```
## Vsphere K3S Machine Config
```yaml
vmwarevsphereMachineConfigs:
    datacenter: "/<datacenter>"                                 #required 
    hostSystem: "/<datacenter>/path-to-host"                    #required
    datastoreURL: "/datastore.URL"                              #required 
    datastore: ""/<datacenter>/path-to-datastore"               #required 
    folder: "/<datacenter>/path-to-vm-folder"                   #required 
    pool: "/<datacenter>/path-to-resource-pool"                 #required 
    vmwarevsphereMachineConfig:
    - cfgparam: ["disk.enableUUID=TRUE"]                        #required
      cloudConfig: "#cloud-config\n\n"
      customAttribute: []
      tag: []
      roles: ["etcd","controlplane",worker]
      creationType: "template"                                  #required
      os: "linux"                                               #required
      cloneFrom: "/<datacenter>/path-to-linux-image"            #required(linux templates only)
      cloneFromWindows: "/<datacenter>/path-to-windows-image"   #required(windows templates only)
      contentLibrary: ""                                        
      datastoreCluster: ""
      network: ["/<datacenter>/path-to-vm-network"]             #required
      sshUser: ""                                               #required
      sshPassword: ""                                           
      sshUserGroup: ""
      sshPort: "22"
      cpuCount: "4"
      diskSize: "40000"
      memorySize: "8192"
```

These tests utilize Go build tags. Due to this, see the below examples on how to run the node driver tests:

`gotestsum --format standard-verbose --packages=github.com/rancher/rancher/tests/v2/validation/provisioning/k3s --junitfile results.xml -- -timeout=60m -tags=validation -v -run "TestK3SProvisioningTestSuite/TestProvisioningK3SCluster"` \
`gotestsum --format standard-verbose --packages=github.com/rancher/rancher/tests/v2/validation/provisioning/k3s --junitfile results.xml -- -timeout=60m -tags=validation -v -run "TestK3SProvisioningTestSuite/TestProvisioningK3SClusterDynamicInput"`

If the specified test passes immediately without warning, try adding the `-count=1` flag to get around this issue. This will avoid previous results from interfering with the new test run.

## Custom Cluster
For custom clusters, no machineConfig or credentials are needed. Currently only supported for ec2.

Dependencies:
* **Ensure you have nodeProviders in provisioningInput**
* make sure that all roles are entered at least once
* windows pool(s) should always be last in the config
```yaml
  awsEC2Configs:
  region: "us-east-2"
  awsSecretAccessKey: ""
  awsAccessKeyID: ""
  awsEC2Config:
    - instanceType: "t3a.medium"
      awsRegionAZ: ""
      awsAMI: ""
      awsSecurityGroups: [""]
      awsSSHKeyName: ""
      awsCICDInstanceTag: "rancher-validation"
      awsIAMProfile: ""
      awsUser: "ubuntu"
      volumeSize: 50
      roles: ["etcd", "controlplane"]
    - instanceType: "t3a.medium"
      awsRegionAZ: ""
      awsAMI: ""
      awsSecurityGroups: [""]
      awsSSHKeyName: ""
      awsCICDInstanceTag: "rancher-validation"
      awsIAMProfile: ""
      awsUser: "ubuntu"
      volumeSize: 50
      roles: ["worker"]
    - instanceType: "t3a.xlarge"
      awsAMI: ""
      awsSecurityGroups: [""]
      awsSSHKeyName: ""
      awsCICDInstanceTag: "rancher-validation"
      awsUser: "Administrator"
      volumeSize: 50
      roles: ["windows"]
```

These tests utilize Go build tags. Due to this, see the below examples on how to run the custom cluster tests:

`gotestsum --format standard-verbose --packages=github.com/rancher/rancher/tests/v2/validation/provisioning/k3s --junitfile results.xml -- -timeout=60m -tags=validation -v -run "TestCustomClusterK3SProvisioningTestSuite/TestProvisioningK3SCustomCluster"` \
`gotestsum --format standard-verbose --packages=github.com/rancher/rancher/tests/v2/validation/provisioning/k3s --junitfile results.xml -- -timeout=60m -tags=validation -v -run "TestCustomClusterK3SProvisioningTestSuite/TestProvisioningK3SCustomClusterDynamicInput"`

If the specified test passes immediately without warning, try adding the `-count=1` flag to get around this issue. This will avoid previous results from interfering with the new test run.

## Template Test

Dependencies:
* [Cloud Credential](#cloud-credentials)
* Make sure the template provider matches the credentials.
```yaml
templateTest:
  repo:
    metadata:
      name: "demo"
    spec:
      gitRepo: "https://github.com/<forked repo>/cluster-template-examples.git"
      gitBranch: main
      insecureSkipTLSVerify: true
  templateProvider: "aws"
  templateName: "cluster-template"
```

`gotestsum --format standard-verbose --packages=github.com/rancher/rancher/tests/v2/validation/provisioning/rke2 --junitfile results.xml -- -timeout=60m -tags=validation -v -run "TestClusterTemplateTestSuite/TestProvisionK3sTemplateCluster`

## Static Test Cases
In an effort to have uniform testing across our internal QA test case reporter, there are specific test cases that are put into their respective test files. This section highlights those test cases.

### PSACT
These test cases cover the following PSACT values as both an admin and standard user:
1. `rancher-privileged`
2. `rancher-restricted`
3. `rancher-baseline`

See an example YAML below:

```yaml
rancher:
  host: "<rancher server url>"
  adminToken: "<rancher admin bearer token>"
  cleanup: false
  clusterName: "<provided cluster name>"
  insecure: true
provisioningInput:
  k3sKubernetesVersion: ["v1.27.10+k3s2"]
  cni: ["calico"]
  providers: ["linode"]
  nodeProviders: ["ec2"]
linodeCredentials:
   token: ""
linodeMachineConfigs:
  region: "us-west"
  linodeMachineConfig:
  - roles: ["etcd", "controlplane", "worker"]
    authorizedUsers: ""
    createPrivateIp: true
    dockerPort: "2376"
    image: "linode/ubuntu22.04"
    instanceType: "g6-standard-8"
    region: "us-west"
    rootPass: ""
    sshPort: "22"
    sshUser: ""
    stackscript: ""
    stackscriptData: ""
    swapSize: "512"
    tags: ""
    uaPrefix: "Rancher"
```

These tests utilize Go build tags. Due to this, see the below examples on how to run the tests:

`gotestsum --format standard-verbose --packages=github.com/rancher/rancher/tests/v2/validation/provisioning/k3s --junitfile results.xml -- -timeout=60m -tags=validation -v -run "TestK3SPSACTTestSuite$"`

### Hardened Custom Cluster
This will provision a hardened custom cluster that runs across the following CIS scan profiles:
1. `k3s-cis-1.8-profile-hardened`
2. `k3s-cis-1.8-profile-permissive`

You would use the same config that you setup for a custom cluster to run this test. Plese reference this [section](#custom-cluster). It also important to note that the machines that you select has `sudo` capabilities. The tests utilize `sudo`, so this can cause issues if there is no `sudo` present on the machine.

These tests utilize Go build tags. Due to this, see the below examples on how to run the tests:

`gotestsum --format standard-verbose --packages=github.com/rancher/rancher/tests/v2/validation/provisioning/k3s --junitfile results.xml -- -timeout=60m -tags=validation -v -run "TestHardenedK3SClusterProvisioningTestSuite$"`

## Advanced Settings
This encapsulates any other setting that is applied in the cluster.spec. Currently we have support for:
* cluster agent customization 
* fleet agent customization

Please read up on general k8s to get an idea of correct formatting for:
* resource requests
* resource limits
* node affinity
* tolerations

```yaml
advancedOptions:
  clusterAgent: # change this to fleetAgent for fleet agent
    appendTolerations:
    - key: "Testkey"
      value: "testValue"
      effect: "NoSchedule"
    overrideResourceRequirements:
      limits:
        cpu: "750m"
        memory: "500Mi"
      requests:
        cpu: "250m"
        memory: "250Mi"
      overrideAffinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - preference:
                matchExpressions:
                  - key: "cattle.io/cluster-agent"
                    operator: "In"
                    values:
                      - "true"
              weight: 1
```