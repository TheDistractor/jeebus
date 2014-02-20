ng = angular.module 'myApp'

ng.value 'appInfo',
  name: 'JeeBus'
  version: '0.3.0'
  home: 'https://github.com/jcw/jeebus'

ng.provider 'navbar', ->
  navs = []
  add: (route, title, weight = 50) ->
    navs.push { route, title, weight }
  del: (route) ->
    navs = navs.filter (x) -> x.route isnt route
  $get: ->
    navs.sort (a, b) -> a.weight - b.weight
  
ng.config ($urlRouterProvider, $locationProvider) ->
  $urlRouterProvider.otherwise '/'
  $locationProvider.html5Mode true
  
ng.run ($rootScope, appInfo) ->
  $rootScope.appInfo = appInfo
  $rootScope.shared = {}
  
ng.controller 'NavCtrl', ($scope, navbar) ->
  $scope.navbar = navbar

ng.directive 'appVersion', (appInfo) ->
  (scope, elm, attrs) ->
    elm.text appInfo.version

#------------------------------------------------------------------------------
# HouseMon-specific setup to connect on startup and define a new "HM" service.
# 
# ng.run (jeebus) ->
#   jeebus.connect 'housemon', 3335
# 
# ng.factory 'HM', (jeebus) ->
#   # For the calls below:
#   #  - if more than one key is specified, they are joined with slashes
#   #  - do not include a slash at the start or end of any key argument
#   
#   keyAsPath = (key) ->
#     return '/'  if key.length is 0
#     "/#{key.join '/'}/"
#     # "/#{['hm'].concat(key).join '/'}/"
#   
#   # Get the sub-keys under a certain path in the host database as a promise.
#   # This only goes one level deep, i.e. a flat list of immediate sub-keys.
#   keys: (key...) ->
#     jeebus.rpc 'db-keys', keyAsPath(key)
#   
#   # Get a key's value from the host database, returned as a promise.
#   get: (key...) ->
#     jeebus.rpc 'db-get', keyAsPath(key)
# 
#   # Set a key/value pair in the host database.
#   # If value is the empty string or null, the key will be deleted.
#   set: (key..., value) ->
#     jeebus.store keyAsPath(key), value
#     @
# 
#   # Attach to a key prefix, returns the tracked model for that prefix.
#   attach: (key...) ->
#     jeebus.attach keyAsPath(key)
# 
#   # Undo the effects of attaching, i.e. stop updating the model entries.
#   detach: (key...) ->
#     jeebus.detach keyAsPath(key)
#     @
