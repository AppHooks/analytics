describe("Adding new service", function(){
    describe("Mixpanel", function(){
        it("should have key field", function(){
            document.querySelector("select[name='service']").value = "mixpanel";
            expect(document.querySelector("input[name='key']")).to.exist;
            expect(document.querySelector("label[for='key']")).to.exist;
        });
    });
    describe("GA", function(){
        it("should have tracking ID field", function(){
            document.querySelector("select[name='service']").value = "ga";
            expect(document.querySelector("input[name='tracking_id']")).to.exist;
            expect(document.querySelector("label[for='tracking_id']")).to.exist;
        });
    });
    describe("Parse", function(){
        it("should have application id and api key", function(){
            document.querySelector("select[name='service']").value = "parse";
            expect(document.querySelector("input[name='application_id']")).to.exist;
            expect(document.querySelector("input[name='api_key']")).to.exist;
            expect(document.querySelector("label[for='application_id']")).to.exist;
            expect(document.querySelector("label[for='api_key']")).to.exist;
        });
    });

});
