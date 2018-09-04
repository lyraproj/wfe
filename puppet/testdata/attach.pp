



type Resource = Object[{
  attributes => {
    title => String
  }
}]

type Genesis::Resource = Resource {
  attributes => {
    ensure => Enum[absent, present],
    region => String,
    tags => Hash[String,String]
  }
}

type Genesis::Aws::Vpc = Genesis::Resource{
  attributes => {
    vpc_id => { type => Optional[String], value => 'FAKED_VPC_ID' },
    cidr_block => String,
    enable_dns_hostnames => Boolean,
    enable_dns_support => Boolean
  }
}

type Genesis::Aws::Subnet = Genesis::Resource{
  attributes => {
    subnet_id => { type => Optional[String], value => 'FAKED_SUBNET_ID' },
    vpc_id => String,
    cidr_block => String,
    map_public_ip_on_launch => Boolean
  }
}

type Genesis::Aws::InternetGateway = Genesis::Resource{
  attributes => {
    internet_gateway_id => { type => Optional[String], value => 'FAKED_GATEWAY_ID' }
  }
}

actor attach(String $region = lookup('aws.region'), Hash[String,String] $tags = lookup('aws.tags')) >> Struct[
    vpc_id => String,
    subnet_id => String,
    internet_gateway_id => String
  ] {
  action vpc(Genesis::Context $ctx, String $region, Hash[String,String] $tags) >> Struct[vpc_id => String, subnet_id => String] {
    $vpc = $ctx.resource(Genesis::Aws::Vpc(
      title => 'nyx-attachinternetgateway-test',
      ensure => present,
      region => $region,
      cidr_block => '192.168.0.0/16',
      tags => $tags,
      enable_dns_hostnames => true,
      enable_dns_support => true
      ))
    $ctx.notice("Created VPC: ${vpc.vpc_id}")

    $subnet = $ctx.resource(Genesis::Aws::Subnet(
      title => 'nyx-attachinternetgateway-test',
      ensure => present,
      region => $region,
      vpc_id => $vpc.vpc_id,
      cidr_block => '192.168.1.0/24',
      tags => $tags,
      map_public_ip_on_launch => true
    ))
    $ctx.notice("Created Subnet: ${subnet.subnet_id}")

    { vpc_id => $vpc.vpc_id, subnet_id => $subnet.subnet_id }
  }

  action gw(Genesis::Context $ctx, String $region, Hash[String,String] $tags) >> Struct[internet_gateway_id => String] {
    $result = $ctx.resource(Genesis::Aws::InternetGateway(
      title => 'nyx-attachinternetgateway-test',
      ensure => present,
      region => $region,
      tags => $tags,
    ))
    $ctx.notice("Created Internet Gateway: ${result.internet_gateway_id}")

    { internet_gateway_id => $result.internet_gateway_id }
  }
}
