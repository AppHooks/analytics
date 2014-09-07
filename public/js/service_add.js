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
    $scope.config = {}

    $scope.changeService = function() {
      $scope.config = {}
    }

    $scope.add = function() {
      $http.post('/services/add', {
        'type': $scope.selected.value,
        'name': $scope.name,
        'config': $scope.config
      })
    }
  }])
