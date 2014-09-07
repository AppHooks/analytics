'use strict';

angular.module('analytics', [])
  .config(['$interpolateProvider', function($interpolate) {
    $interpolate.startSymbol('<?');
    $interpolate.endSymbol('?>');
  }])
  .controller('AddServiceCtrl', ['$scope', '$http', function($scope, $http){
    $scope.services = [
      { name: 'Mixpanel', value: 'mixpanel' }, 
      { name: 'Google Analytics', value: 'ga' },
      { name: 'Parse', value: 'parse' },
    ];
    $scope.selected = $scope.services[0]
    $scope.add = function() {
      $http.post('/services/add', {
        "type": "mixpanel",
        "name": "mixpanel1",
        "config": {
          "key": "key",
          "secret": "secret"
        }
      })
    }
  }])
