POST http://localhost:8080/v1/sequences
Content-Type: application/json
{
  "name": "Test Sequence",
  "openTrackingEnabled": true,
  "clickTrackingEnabled": true,
  "steps": [
    {
      "emailSubject": "Test Email Subject 1",
      "emailContent": "Test Email Content 1",
      "daysAfterPreviousStep": 0
    },
    {
      "emailSubject": "Test Email Subject 2",
      "emailContent": "Test Email Content 2",
      "daysAfterPreviousStep": 2
    },
    {
      "emailSubject": "Test Email Subject 3",
      "emailContent": "Test Email Content 3",
      "daysAfterPreviousStep": 3
    }
  ]
}
HTTP 201

[Captures]
sequence-id: jsonpath "$['id']"
first-sequence-step-id: jsonpath "$['steps'][0]['id']"
second-sequence-step-id: jsonpath "$['steps'][1]['id']"
third-sequence-step-id: jsonpath "$['steps'][2]['id']"

###

PUT http://localhost:8080/v1/sequences/{{sequence-id}}/steps/{{first-sequence-step-id}}
Content-Type: application/json
{
  "emailSubject": "Updated Test Email Subject",
  "emailContent": "Updated Test Email Content"
}
HTTP 200

###

DELETE http://localhost:8080/v1/sequences/{{sequence-id}}/steps/{{second-sequence-step-id}}
HTTP 204

###

PUT http://localhost:8080/v1/sequences/{{sequence-id}}
Content-Type: application/json
{
  "openTrackingEnabled": false,
  "clickTrackingEnabled": false
}
HTTP 200

[Asserts]
jsonpath "$['id']" == "{{sequence-id}}"
jsonpath "$['name']" == "Test Sequence"
jsonpath "$['openTrackingEnabled']" == false
jsonpath "$['clickTrackingEnabled']" == false
jsonpath "$['steps']" count == 2
jsonpath "$['steps'][0]['id']" == "{{first-sequence-step-id}}"
jsonpath "$['steps'][0]['emailSubject']" == "Updated Test Email Subject"
jsonpath "$['steps'][0]['emailContent']" == "Updated Test Email Content"
jsonpath "$['steps'][0]['daysAfterPreviousStep']" == 0
jsonpath "$['steps'][1]['id']" == "{{third-sequence-step-id}}"
jsonpath "$['steps'][1]['emailSubject']" == "Test Email Subject 3"
jsonpath "$['steps'][1]['emailContent']" == "Test Email Content 3"
jsonpath "$['steps'][1]['daysAfterPreviousStep']" == 3