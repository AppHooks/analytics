describe("Adding new service", function(){
  
  before(function() {
    casper.start('http://localhost:3000', function () {
      this.fill('form', {
        'email': 'admin@email.com',
        'password': 'password'
      }, true)
    });
  })

  describe("Add GA", function(){
    it ("should have tracking_id field", function(){
      casper.thenOpenAndEvaluate('http://localhost:3000/services/add.html', function() {
        document.querySelector("select[name='service']").value = "ga";
      })
      .then(function() {
        'input[name="tracking_id"]'.should.be.inDOM.and.be.visible;
        'label[for="tracking_id"]'.should.be.inDOM.and.be.visible;
      });
    })
  })

  describe("Add mixpanel", function(){
    it ("should have key field", function(){
      casper.thenOpenAndEvaluate('http://localhost:3000/services/add.html', function() {
        document.querySelector("select[name='service']").value = "mixpanel";
      })
      .then(function() {
        'input[name="key"]'.should.be.inDOM.and.be.visible;
        'label[for="key"]'.should.be.inDOM.and.be.visible;
      });
    })
  })

  describe("Add parse", function(){
    it ("should have application_id and api_key field", function(){
      casper.thenOpenAndEvaluate('http://localhost:3000/services/add.html', function() {
        document.querySelector("select[name='service']").value = "parse";
      })
      .then(function() {
        'input[name="application_id"]'.should.be.inDOM.and.be.visible;
        'label[for="application_id"]'.should.be.inDOM.and.be.visible;
        'input[name="api_key"]'.should.be.inDOM.and.be.visible;
        'label[for="api_key"]'.should.be.inDOM.and.be.visible;
      });
    })
  })
});
