# Countries Dashboard Service - Assignment 2

## Description
This project is our group submission to the second assignment in PROG2005.

In this assignment we have developed a REST web application in Golang that provides the client with the ability to configure information dashboards about countries, that are dynamically populated when requested. They are populated based on external services, and includes a simple notification service that can listen to specific events. The service manage state using databases, and webhooks as notification feature. 

The external services used:
* *REST Countries API* (instance hosted)
  * Endpoint: http://129.241.150.113:8080/v3.1
  * Documentation: http://129.241.150.113:8080/
* *Open-Meteo APIs* (hosted externally)
  * Documentation: https://open-meteo.com/en/features#available-apis
* *Currency API*
  * Endpoint: http://129.241.150.113:9090/currency/
  * Documentation: http://129.241.150.113:9090/

The final web service is deployed in our local OpenStack instance SkyHigh. 

## Endpoints

Our web service have four resource root paths: 

```
/dashboard/v1/registrations/
/dashboard/v1/dashboards/
/dashboard/v1/notifications/
/dashboard/v1/status/
```


## Endpoint 'Registrations':

The initial endpoint focuses on the management of dashboard configurations that can later be used via the `dashboards` endpoint.

### Register new dashboard configuration

Manages the registration of new dashboard configurations that indicate which information is to be shown for registered dashboards. This includes weather, country and currency exchange information.

**Request** **(POST)** 

```
Method: POST
Path: /dashboard/v1/registrations/
Content type: application/json
```

Body (exemplary code):
```
{
   "country": "Norway",                                     // Indicates country name (alternatively to ISO code, i.e., country name can be empty if ISO code field is filled and vice versa)
   "isoCode": "NO",                                         // Indicates two-letter ISO code for country (alternatively to country name)
   "features": {
                  "temperature": true,                      // Indicates whether temperature in degree Celsius is shown
                  "precipitation": true,                    // Indicates whether precipitation (rain, showers and snow) is shown
                  "capital": true,                          // Indicates whether the name of the capital is shown
                  "coordinates": true,                      // Indicates whether country coordinates are shown
                  "population": true,                       // Indicates whether population is shown
                  "area": true,                             // Indicates whether land area size is shown
                  "targetCurrencies": ["EUR", "USD", "SEK"] // Indicates which exchange rates (to target currencies) relative to the base currency of the registered country (in this case NOK for Norway) are shown
               }
}
```

**Response**

The response stores the configuration on the server and returns the associated ID. In the example below, it is the ID `1`. Responses show be encoded in the above-mentioned JSON format, with the `lastChange` field highlighting the last change to the configuration (including updates via `PUT`)

* Content type: `application/json`
* Status code: Appropriate error code.

Body (exemplary code for registered configuration):
```
{
    "id": 1
    "lastChange": "2024-02-29 12:31"
}
```

### View a **specific registered dashboard configuration**

**Request (GET)**

The following shows a request for an individual configuration identified by its ID.

```
Method: GET
Path: /dashboard/v1/registrations/{id}
```

* `id` is the ID associated with the specific configuration.

Example request: ```/dashboard/v1/registrations/1``` 

**Response**

* Content type: `application/json`
* Status code: Appropriate error code. 

Body (exemplary code):
```
{
   "id": 1,
   "country": "Norway",
   "isoCode": "NO",
   "features": {
                  "temperature": true,
                  "precipitation": true,
                  "capital": true,
                  "coordinates": true,
                  "population": true,
                  "area": false,
                  "targetCurrencies": ["EUR", "USD", "SEK"]
               },
    "lastChange": "20240229 14:07"
}
```

### View **all registered dashboard configurations**

**Request (GET)**

A `GET` request to the endpoint return all registered configurations including IDs and timestamps of last change.

```
Method: GET
Path: /dashboard/v1/registrations/
```

**Response**

* Content type: `application/json`
* Status code: Appropriate error code.

Body (exemplary code):
```
[
   {
      "id": 1,
      "country": "Norway",
      "isoCode": "NO",
      "features": {
                     "temperature": true,
                     "precipitation": true,
                     "capital": true,
                     "coordinates": true,
                     "population": true,
                     "area": false,
                     "targetCurrencies": ["EUR", "USD", "SEK"]
                  }, 
      "lastChange": "20240229 14:07"
   },
   {
      "id": 2,
      "country": "Denmark",
      "isoCode": "DK",
      "features": {
                     "temperature": false,
                     "precipitation": true,
                     "capital": true,
                     "coordinates": true,
                     "population": false,
                     "area": true,
                     "targetCurrencies": ["NOK", "MYR", "JPY", "EUR"]
                  },
       "lastChange": "20240224 08:27"
   },
   ...
]
```

The response return a collection of return all stored configurations.


### Replace specific registered dashboard configurations


**Request (PUT) & (PATCH)**

This request allows updating an individual configuration identified by its ID. Using either PUT or PATCH, you can modify the configuration, which also updates the associated timestamp (lastChange). The PUT method replaces the entire configuration with the new data provided in the request body, while the PATCH method applies partial updates to the configuration, allowing for more granular changes without affecting the entire configuration at once

```
Method: PUT / PATCH
Path: /dashboard/v1/registrations/{id}
```

* `id` is the ID associated with the specific configuration.

Example request PUT: ```/dashboard/v1/registrations/1``` 

Body (exemplary code):
```
{
   "country": "Norway",
   "isoCode": "NO",
   "features": {
                  "temperature": false, // this value is to be changed
                  "precipitation": true,
                  "capital": true,
                  "coordinates": true, 
                  "population": true,
                  "area": false,
                  "targetCurrencies": ["EUR", "SEK"] // this value is to be changed
               }
}
```

**Response**

This is the response to the change request.

* Status code: Appropriate error code.
* Body: empty

### Delete a **specific registered dashboard configuration**

**Request (DELETE)**

The following shows a request for deletion of an individual configuration identified by its ID.

```
Method: DELETE
Path: /dashboard/v1/registrations/{id}
```

* `id` is the ID associated with the specific configuration.

Example request: ```/dashboard/v1/registrations/1``` 

**Response**

This is the response to the delete request.

* Status code: Appropriate error code.
* Body: Message it has been deleted

## Endpoint 'Dashboards':

This endpoint can be used to retrieve the populated dashboards.

**Request (GET)** 

Retrieving a specific populated dashboard
The following shows a request for an individual dashboard identified by its ID.

```
Method: GET
Path: /dashboard/v1/dashboards/{id}
```

* `id` is the ID associated with the specific configuration.

Example request: ```/dashboard/v1/dashboards/1``` 

**Response**

* Content type: `application/json`

Body (exemplary code):
```
{
   "country": "Norway",
   "isoCode": "NO",
   "features": {
                  "temperature": -1.2,                       // Mean temperature across all forecasted temperature values for country's coordinates
                  "precipitation": 0.80,                     // Mean precipitation across all returned precipitation values
                  "capital": "Oslo",                         // Capital: Where multiple values exist, take the first
                  "coordinates": {                           // Those are the country geocoordinates
                                    "latitude": 62.0,
                                    "longitude": 10.0
                                 },
                  "population": 5379475,
                  "area": 323802.0,
                  "targetCurrencies": {
                                         "EUR": 0.087701435,  // this is the current NOK to EUR exchange rate (where multiple currencies exist for a given country, take the first)
                                         "USD": 0.095184741, 
                                         "SEK": 0.97827275
                                       }
               },
    "lastRetrieval": "20240229 18:15" // this should be the current time (i.e., the time of retrieval)
}
```

## Endpoint 'Notifications': Managing webhooks for event notifications

The users can register webhooks that are triggered by the service based on specified events, specifically if a new configuration is created, changed or deleted. Users can also register for invocation events, i.e., when a dashboard for a given country is invoked. Users can register multiple webhooks, and they are persistently stored.

### Registration of Webhook

**Request (POST)**

```
Method: POST
Path: /dashboard/v1/notifications/
Content type: application/json
```

The body contains 
 * the URL to be triggered upon event (the service that should be invoked)
 * the country for which the trigger applies (if empty, it applies to any invocation)
 * Events: 
   * `REGISTER` - webhook is invoked if a new configuration is registered
   * `CHANGE` - webhook is invoked if configuration is modified
   * `DELETE` - webhook is invoked if configuration is deleted
   * `INVOKE` - webhook is invoked if dashboard is retrieved (i.e., populated with values)

Body (Exemplary message based on schema):
```
{
   "url": "https://localhost:8080/client/",  // URL to be invoked when event occurs
   "country": "NO",                          // Country that is registered, or empty if all countries
   "event": "INVOKE"                         // Event on which it is invoked
}
```
**Response**

* Content type: `application/json`

Body (Exemplary message based on schema):
```
{
    "id": "OIdksUDwveiwe"
}
```

### Deletion of Webhook

**Request (DELETE)**

```
Method: DELETE
Path: /dashboard/v1/notifications/{id}
```

* {id} is the ID returned during the webhook registration

**Response**

Returns success as a http.StatusNoContent  

### View *specific registered* webhook

**Request (GET)**

```
Method: GET
Path: /dashboard/v1/notifications/{id}
```
* `{id}` is the ID for the webhook registration

**Response**

* Content type: `application/json`

Body (Exemplary message based on schema):
```
{
   "id": "OIdksUDwveiwe",
   "url": "https://localhost:8080/client/",
   "country": "NO",
   "event": "INVOKE"
}

```

### View *all registered* webhooks

**Request (GET)**

```
Method: GET
Path: /dashboard/v1/notifications/
```

**Response**

The response is a collection of all registered webhooks.

* Content type: `application/json`

Body (Exemplary message based on schema):
```
[
   {
      "id": "OIdksUDwveiwe",
      "url": "https://localhost:8080/client/",
      "country": "NO",
      "event": "INVOKE"
   },
   {
      "webhook_id": "DiSoisivucios",
      "url": "https://localhost:8081/anotherClient/",
      "country": "",                                 // field can also be omitted if registered for all countries
      "event": "REGISTER"
   },
   ...
]
```

### Webhook Invocation (upon trigger)

When a webhook is triggered, it sends information as follows. Where multiple webhooks are triggered, the information is sent separately. 

```
Method: POST
Path: <url specified in the corresponding webhook registration>
Content type: application/json
```

Body (Exemplary message based on schema):
```
{
   "id": "OIdksUDwveiwe",
   "country": "NO",
   "event": "INVOKE",
   "time": "20240223 06:23"      // time at which the event occurred
}
```


## Endpoint 'Status'
This endpoint is monitoring service availability, indicating availability on services this service depends on reporting appropriate error codes. With the addition of information about number of webhooks and uptime of the service. 

```
Method: GET
Path: dashboard/v1/status/
```

**Response**

* Content type: `application/json`
* Status code: appropriate error code. 

Body:
```
{
   "countries_api": <http status code for *REST Countries API*>,
   "meteo_api": <http status code for *Meteo API*>, 
   "currency_api": <http status code for *Currency API*>,
   "notification_db": <http status code for *Notification database*>,
   ...
   "webhooks": <number of registered webhooks>,
   "version": "v1",
   "uptime": <time in seconds from the last service restart>
}
```

## Additional requirements

* All endpoints is *tested using automated testing facilities (unit tests, httptest package)*. 
  * Including the stubbing of the third-party endpoints to ensure test reliability (removing dependency on external services). 
* Used Firebase to minimize invocation on third-party libraries. 


## Test 
To run the tests run command in root folder:  go test -v ./.... Here you will see what tests are run, and if they pass or fail. 

There are test for functions that do not directly access firestore, or parts of functions that do not access firestore. With mocking of firestore, it is possible to make the % of tests much higher, but due to time and problems mocking firestore (and to our understanding, it is not in our curriculum to be able to mock firestore) these are not tested. 


## Deployment

The service is to be deployed on an IaaS solution OpenStack using Docker. 

URL to deployed service on SkyHigh: http://10.212.169.134:8080/ (requires VPN)

To run this code yourself, you need to connect it to your own firestore, by making databases there (named Dashboard and webhooks), and put your firestore key in the same folder as main.go