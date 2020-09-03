# cloudeos_router_status

The `cloudeos_router_status` resource should be created after a CloudEOS router has been deployed. It sends all the information
about the deployed CloudEOS router to CVaaS. Unlike `cloudeos_router_config` which takes minimal input about how the
CloudEOS router should be deployed, `cloudeos_router_status` provides detailed deployment information after the router
is deployed.

## Example Usage

```hcl
resource "azurerm_resource_group" "rg" {
  name       = "example-rg"
  location   = "westus2"
}

resource "azurerm_virtual_network" "vnet" {
  name                = "example-vnet"
  address_space       = ["10.0.0.0/16"]
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
}

resource "azurerm_subnet" "internal" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.rg.name
  address_prefix       = "10.0.2.0/24"
}

resource "azurerm_public_ip" "publicip" {
  name                = "cloudeos-pip"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  allocation_method   = "Static"
  zones               = [2]
}

resource "azurerm_subnet" "public" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.rg.name
  address_prefix       = "10.0.1.0/24"
}

resource "azurerm_network_interface" "internalIntf" {
  name                = "internalIntf"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  ip_configuration {
    name                          = "internalIp"
    subnet_id                     = azurerm_subnet.internal.id
    private_ip_address            = "10.0.2.101"
  }
}

resource "azurerm_network_interface" "publicIntf" {
  name                = "publicIntf"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  ip_configuration {
    name                          = "publicIp"
    subnet_id                     = azurerm_subnet.public.id
    private_ip_address            = "10.0.1.101"
    public_ip_address_id          = azurerm_public_ip.publicip.id
  }
}

resource "cloudeos_router_config" "cloudeos" {
  cloud_provider = "azure"
  topology_name = cloudeos_topology.topology.topology_name
  role = "CloudEdge"
  cnps = "dev"
  vpc_id = azurerm_virtual_network.vnet.id
  region = azurerm_resource_group.rg.location
  is_rr = false
  intf_name = ["publicIntf", "internalIntf"]
  intf_private_ip = [azurerm_network_interface.publicIntf.private_ip_address,
                     azurerm_network_interface.internalIntf.private_ip_address]
  intf_type = ["public", "internal"]
}

data "template_file" "user_data_specific" {
  template = file("file.txt")
  vars = {
    bootstrap_cfg = cloudeos_router_config.router[0].bootstrap_cfg
  }
}

resource "azurerm_virtual_machine" "cloudeosVm" {
  name                          = "example-cloudeos"
  location                      = azurerm_resource_group.rg.location
  resource_group_name           = azurerm_resource_group.rg.name
  vm_size                       = "Standard_D4_v2"
  primary_network_interface_id  = azurerm_network_interface.publicIntf.id
  network_interface_ids         = [azurerm_network_interface.publicIntf.id, azurerm_network_interface.internalIntf.id]

  storage_image_reference {
    publisher = "arista-networks"
    offer     = "cloudeos-router-payg"
    sku       = "cloudeos-4_24_0-payg"
    version   = "4.24.01"
  }

  storage_os_disk {
    name              = "cloudeos-disk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "example-cloudeos"
    admin_username = "testadmin"
    admin_password = "Password1234!"
    custom_data = data.template_file.user_data_specific[0].rendered
  }
  os_profile_linux_config {
    disable_password_authentication = false
  }
}

resource "cloudeos_router_status" "cloudeos" {
  cloud_provider = "azure"
  cv_container   = "CloudEdge"
  cnps = "dev"
  vpc_id = azurerm_virtual_network.vnet.id
  instance_id = azurerm_virtual_machine.cloueosVm.id
  instance_type = azurerm_virtual_machine.cloueosVm.instance_type
  region = "westus2"
  primary_network_interface_id = azurerm_network_interface.publicIntf.id
  public_ip = azurerm_public_ip.publicip.ip_address
  intf_name = cloudeos_router_config.cloudeos.intf_name
  intf_id =  [azurerm_network_interface.publicIntf.id, azurerm_network_interface.internalIntf.id]
  intf_private_ip = cloudeos_router_config.cloudeos.intf_private_ip
  intf_subnet_id = [azurerm_subnet.public.id, azurerm_subnet.internal.id]
  intf_type = cloudeos_router_config.cloudeos.intf_type
  tf_id = cloudeos_router_config.cloudeos.tf_id
  is_rr = "false"
}
```

## Argument Reference

* `cloud_provider` - (Required) CloudProvider type. Supports only aws or azure.
* `instance_type` - (Required) Instance ID of deployed CloudEOS.
* `intf_name` - (Required) List of interface names for the routers.
* `intf_id` - (Required) List of interface IDs attached to the routers.
* `intf_private_ip` - (Required) List of private IPs attached to the interfaces.
* `intf_subnet_id` - (Required) List of subnet IDs of interfaces.
* `intf_type` - (Required) List of interface types. Values supported : public, internal, private.
* `region` - (Required) Region of deployment.
* `cv_container` - (Optional) Container in CVaaS to which the router will be added to.
* `vpc_id` - (Optional) VPC/VNET ID of the VPC in which the CloudEOS is deployed in.
* `rg_name` - (Optional) Resource group name, only for Azure.
* `rg_location` - (Optional) Resource group location, only for Azure.
* `tags` - (Optional) A mapping of tags to assign to the resource.
* `availability_zone` - (Optional) Availability Zone in which the router is deployed in.
* `primary_network_interface_id` - (Optional)
* `availability_set_id` - (Optional) Availability Set.
* `public_ip` - (Optional) Public IP of interface.
* `private_rt_table_ids` - (Optional) List of private interface route table IDs.
* `internal_rt_table_ids` - (Optional) List of internal interface route table IDs.
* `public_rt_table_ids` - (Optional) List of public route table IDs.
* `ha_name` - (Optional) Cloud HA pair name.
* `cnps` - (Optional) Cloud Network Private Segments ( VRF name )
* `is_rr` - (Optional) true if this CloudEOS acts as a Route Reflector.

## Attributes Reference

In addition to Arguments listed above - the following Attributes are exported

* `ID` - The ID of cloudeos_router_status Resource.

## Timeouts

* `delete` - (Defaults to 10 minutes) Used when deleting the cloudeos_status Resource.