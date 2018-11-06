

type Genesis::Aws::Tagspecification = Object {
  attributes => {
    resource_type => String, tags => Hash[String, String]
  }
}

type Resource = Object{
}

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

type Genesis::Aws::Instance = Genesis::Resource{
  attributes => {
    instance_id => { type => Optional[String], value => 'FAKED_INSTANCE_ID' },
    instance_type => String,
    image_id => String,
    key_name => String,
    public_ip => { type => Optional[String], value => 'FAKED_PUBLIC_IP' },
    private_ip => { type => Optional[String], value => 'FAKED_PRIVATE_IP' },
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

workflow attach {
  typespace => 'genesis::aws',
  input => (
    String $region = lookup('aws.region'),
    Hash[String,String] $tags = lookup('aws.tags'),
    String $key_name = lookup('aws.keyname'),
    Integer $ec2_cnt = lookup('aws.instance.count')
  ),
  output => (
    String $vpc_id,
    String $subnet_id,
    String $internet_gateway_id,
    Hash[String, Struct[public_ip => String, private_ip => String]] $nodes
  )
} {
  resource vpc {
    input  => ($region, $tags),
    output => ($vpc_id)
  }{
    ensure => present,
    region => $region,
    cidr_block => '192.168.0.0/16',
    tags => $tags,
    enable_dns_hostnames => true,
    enable_dns_support => true
  }

  resource subnet {
    input  => ($region, $tags, $vpc_id),
    output => ($subnet_id)
  }{
    ensure => present,
    region => $region,
    vpc_id => $vpc_id,
    cidr_block => '192.168.1.0/24',
    tags => $tags,
    map_public_ip_on_launch => true
  }

  resource instance {
    input => ($n, $region, $key_name, $tags),
    output => ($key = instance_id, $value = (public_ip, private_ip))
  } $nodes = times($ec2_cnt) |$n| {
    region => $region,
    ensure => present,
    instance_id => String($n, '%X'),
    image_id => 'ami-f90a4880',
    instance_type => 't2.nano',
    key_name => $key_name,
    tags => $tags
  }

  resource internetgateway {
    input => ($region, $tags),
    output => ($internet_gateway_id)
  } {
    ensure => present,
    region => $region,
    tags   => $tags,
  }
}
