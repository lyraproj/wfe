
function ActorContext() {
  var PROTO_PATH = __dirname + '/../fsmpb';
  var grpc = require('grpc');
  var protoLoader = require('@grpc/proto-loader');

  var packageDefinition = protoLoader.loadSync(
    PROTO_PATH,
    {keepCase: true,
      longs: String,
      enums: String,
      defaults: true,
      oneofs: true,
      includeDirs: ['/../../data-protobuf']
    });

  var protoDescriptor = grpc.loadPackageDefinition(packageDefinition);

// The protoDescriptor object has the full package hierarchy
  var actor = protoDescriptor.actor;

  var getActions = function(call) {

  };

  var invokeAction = function(call) {

  };

  var server = new grpc.Server();
  server.addProtoService(actor.Actor.service, {
    getActions,
    invokeAction
  });

  return server;
}

function test() {
  genesis.addContextFunction('tier1',
    {
      produces: {
        c: 'String',
        d: 'Integer',
      },
      consumes:  {
        a: 'String',
        b: 'Integer',
      },
    },
    function (ctx, a, b) {
      return { c: null, d: null }
    }
  )
}

var v = new ActorContext();