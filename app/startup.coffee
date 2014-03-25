ng = angular.module 'myApp', [
  'ui.router'
  'ngAnimate'
  'mm.foundation'
]

ng.value 'appInfo',
  name: 'JeeBus'
  version: '0.9.0'
  home: 'https://github.com/jcw/jeebus'

ng.run (jeebus) ->
  jeebus.connect 'jeebus'

ng.run ($rootScope, appInfo) ->
  $rootScope.shared = {}
  $rootScope.appInfo = appInfo
  $rootScope.$on 'ws-open', ->
    $rootScope.serverStatus = 'connected'
  $rootScope.$on 'ws-lost', ->
    $rootScope.serverStatus = 'disconnected'
  window.$rootScope = $rootScope # console access, for debugging

ng.directive 'highlightOnChange', ($animate) ->
  (scope, elem, attrs) ->
    scope.$watch attrs.highlightOnChange, ->
      $animate.addClass elem, 'highlight', ->
        attrs.$removeClass 'highlight'
