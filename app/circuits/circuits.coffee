ng = angular.module 'myApp'

ng.config ($stateProvider, navbarProvider) ->
  $stateProvider.state 'circuits',
    url: '/circuits'
    templateUrl: '/circuits/circuits.html'
    controller: circuitsCtrl
  navbarProvider.add '/circuits', 'Circuits', 30

circuitsCtrl = ($scope, jeebus) ->
  $scope.myDefs =
    Pipe:
      name: 'Pipeline'
      width: 160
      icon: '\uf061' # fa-arrow-right
      pins: [
        { name: 'In', dir: 'in' }
        { name: 'Out', dir: 'out' }
      ]
    Printer:
      width: 120
      shade: 'lightblue'
      icon: '\uf02f' # fa-print
      pins: [
        { name: 'In', dir: 'in' }
        { name: 'In2', dir: 'in' }
      ]
      
  $scope.myData =
    gadgets: [
      { id:'g1', x:800, y:70, title:'Gadget One', type: 'Printer' }
      { id:'g2', x:800, y:170, title:'Gadget Two', type: 'Pipe' }
    ]
    wires: [
      { from: "g2.Out", to: "g1.In", cap: 0 }
    ]
    feeds:
      "g1.In": { data: [ "some data", "more data" ], x: 800, y: 270 }
    labels:
      "In": "g2.In"
  
  # setup = ->
  #   jeebus.attach 'circuit'
  #     .on 'sync', ->
  #       $scope.circuits = @rows
  #     
  # setup()  if $scope.serverStatus is 'connected'
  # $scope.$on 'ws-open', setup
