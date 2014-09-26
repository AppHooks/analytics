Analytics
=========

Data collector for sending to 3rd party analytics.

##How to use
Make a POST request to 
  
`http://llun-analytics.herokuapp.com/send/YOUR_KEY`
  
with the following data:
````
{
  "event": "name",
  "data": {
  	"key1": "value1",
  	"key2": "value2"
  }
}
````

##Response
Success:
````
{
  "success": true,
  "services": {
    "Google Analytics": true,
    "Mixpanel": true
  }
}
````
