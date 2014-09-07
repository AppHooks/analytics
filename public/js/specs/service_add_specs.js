'use strict'

describe('Add service controller', function() {

	var scope, httpBackend, createController
  var expect = chai.expect

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

  describe('change service', function() {

    it ('should clear config value', function() {
      var ctrl = createController()

      scope.config = { 'key': 'value' }
      scope.changeService()

      expect(scope.config).to.deep.equal({})

    })

  })

	describe('submit form', function() {

    var ctrl

    beforeEach(function() {
      ctrl = createController()
    })

    it ('should use value from services as service type', function () {
      httpBackend.expectPOST('/services/add', {
        'name': 'other',
        'type': 'ga',
        'config': {
          'tracking_id': 'key'
        }
      }).respond(302, '', { 'location': '/services/list.html' })

      scope.name = 'other'
      scope.selected = scope.services[1]
      scope.config = { 'tracking_id': 'key' }
      scope.add()
      httpBackend.flush()

    })

    it ('should put group configuration to single field', function () {
      httpBackend.expectPOST('/services/add', {
        'name': 'mixpanel',
        'type': 'mixpanel',
        'config': {
          'key': 'mixpanelkey'
        }
      }).respond(302, '', { 'location': '/services/list.html' })

      scope.name = 'mixpanel'
      scope.config = { 'key': 'mixpanelkey' }
      scope.add()
      httpBackend.flush()

    })

	})

})
