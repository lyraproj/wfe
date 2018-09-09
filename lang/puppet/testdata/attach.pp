

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

  # An action is a multi action when the parameter declaration is prefixed by an
  # "iterate" keyword followed by an expression. The expression must evaluate to an
  # integer (iterate this number of times), an array (iterate over these elements),
  # or a hash (iterate over these associations).
  #
  # The output of such an action is a variable named after the action itself. It is
  # a Hash for which each iteration produces an association in the form of a two
  # element tuple representing the key and value.
  #
  # In practice, the iterations will be processed by a worker pool that executes
  # them in parallel and the resulting hash is made available to subsequent actions
  # when all iterations have completed.
  #
  # An integer or array iteration adds one parameter to the beginning of the
  # action's parameter list (value of counter or current element). A hash iteration
  # adds two parameters (key and value of current element).
  #
  # E.g. for an array type prefix expression, this declaration:
  #
  #   action nodes iterate 5 (Integer $e, ...)
  #     >> Tuple[String, Struct[public_ip => String, private_ip => String]] { ... }
  #
  # will result in the output type:
  #
  #   Struct[nodes =>
  #     Hash[String, Struct[String, public_ip => String, private_ip => String]]]
  #
  # For a hash type prefix expression ($nodes assumed to be produced in previous
  # example), this declaration:
  #
  #   action check_nodes iterate $nodes (
  #     String $instance_id,
  #     Struct[public_ip => String, private_ip => String] $node,
  #     ...)
  #       >> Tuple[String, Boolean] { ... }
  #
  # will result in the output type:
  #
  #     Struct[check_nodes => Hash[String, Boolean]]
  #
  # This example creates five instances and produces the variable $nodes which is
  # a hash keyed by the instance_id of the created instances.
  #

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
