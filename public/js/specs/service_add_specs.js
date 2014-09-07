'use strict'

describe('Add service controller', function() {

	var scope, httpBackend, createController

  beforeEach(module('analytics'))
	beforeEach(inject(function($rootScope, $httpBackend, $controller) {
		httpBackend = $httpBackend
		scope = $rootScope.$new()

		createController = function() {
      return $controller('AddServiceCtrl', {
        '$scope': scope
      })
		}
	}))

  afterEach(function() {
    httpBackend.verifyNoOutstandingExpectation();
    httpBackend.verifyNoOutstandingRequest();
  });

	describe('submit form', function() {

    it ('should put group configuration to single field', function () {

      var ctrl = createController()

      httpBackend.expectPOST('/services/add').respond(200, '')
      scope.add()
      httpBackend.flush()

    })

	})

})
