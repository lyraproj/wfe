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
