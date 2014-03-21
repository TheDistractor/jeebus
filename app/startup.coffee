ng = angular.module 'myApp', [
  'ui.router'
  'ngAnimate'
  'mm.foundation'
]

ng.value 'appInfo',
  name: 'JeeBus'
  version: '0.3.0'
  home: 'https://github.com/jcw/jeebus'

ng.run (jeebus) ->
  jeebus.connect 'jeebus'

ng.run ($rootScope, appInfo) ->
  $rootScope.appInfo = appInfo
  $rootScope.shared = {}
