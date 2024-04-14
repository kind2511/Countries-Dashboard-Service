# Countries Dashboard Service

[TOC]

# Overview

In this group assignment, we have develop a REST web application in Golang that provides the client with the ability to configure information dashboards that are dynamically populated when requested. The dashboard configurations are saved in our service in a persistent way, and populated based on external services. It also include a simple notification service that can listen to specific events. The application is dockerized and deployed using an IaaS system. The service manage state using databases, as well as webhooks as notification feature and a different deployment mechanism.

The services used for this purpose are:
* *REST Countries API* (instance hosted)
  * Endpoint: http://129.241.150.113:8080/v3.1
  * Documentation: http://129.241.150.113:8080/
* *Open-Meteo APIs* (hosted externally)
  * Documentation: https://open-meteo.com/en/features#available-apis
* *Currency API*
  * Endpoint: http://129.241.150.113:9090/currency/
  * Documentation: http://129.241.150.113:9090/

The final web service is deployed in our local OpenStack instance SkyHigh. 

# Specification

The implementation of the service API follow this specification, i.e., the schemas (or syntax) of request and response messages, alongside method and status codes should correspond to the ones provided below. Requests and responses are expressed using examples to illustrate the structure in populated messages.


## Endpoints

Our web service have four resource root paths: 

```
/dashboard/v1/registrations/
/dashboard/v1/dashboards/
/dashboard/v1/notifications/
/dashboard/v1/status/
```

The specification has the following conventions for placeholders:

* {value} - *mandatory* value
* {value?} - *optional* value
* {?key=value} - *mandatory* parameter (key-value pair)
* {?key=value?} - *optional* parameter (key-value pair)
* {?key=value*} - one or more optional parameters 

## Endpoint 'Registrations': Registering dashboard configuration

The initial endpoint focuses on the management of dashboard configurations that can later be used via the `dashboards` endpoint.

### Register new dashboard configuration

Manages the registration of new dashboard configurations that indicate which information is to be shown for registered dashboards. This includes weather, country and currency exchange information.

### - Request (POST) 

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

Enables retrieval of a specific registered dashboard configuration.

### - Request (GET)

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

Enables retrieval of all registered dashboard configurations.

### - Request (GET)

A `GET` request to the endpoint should return all registered configurations including IDs and timestamps of last change.

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


### Replace a **specific registered dashboard configuration**

Enables the replacing of specific registered dashboard configurations.

### - Request (PUT) & (PATCH)

The following shows a request for an updated of individual configuration identified by its ID. This update should lead to an update of the configuration and an update of the associated timestamp (`lastChange`).

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

Note that the request neither contains ID in the body (only in the URL), and neither contains the timestamp.  
The PATCH method has also been implemented here. 

**Response**

This is the response to the change request.

* Status code: Appropriate error code.
* Body: empty

### Delete a **specific registered dashboard configuration**

Enabling the deletion of a specific registered dashboard configuration.

### - Request (DELETE)

The following shows a request for deletion of an individual configuration identified by its ID. This update should lead to a deletion of the configuration on the server.

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

## Endpoint 'Dashboards': Retrieve populated dashboard

This endpoint can be used to retrieve the populated dashboards.

### - Request (GET) - Retrieving a **specific populated dashboard**

The following shows a request for an individual dashboard identified by its ID (same as the corresponding configuration ID).

```
Method: GET
Path: /dashboard/v1/dashboards/{id}
```

* `id` is the ID associated with the specific configuration.

Example request: ```/dashboard/v1/dashboards/1``` 

**Response**

* Content type: `application/json`
* Status code: Appropriate error code.

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

As an additional feature, users can register webhooks that are triggered by the service based on specified events, specifically if a new configuration is created, changed or deleted. Users can also register for invocation events, i.e., when a dashboard for a given country is invoked. Users can register multiple webhooks. The registrations should survive a service restart (i.e., be persistently stored).

### Registration of Webhook

### - Request (POST)

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

The response contains the ID for the registration that can be used to see detail information or to delete the webhook registration. The format of the ID is not prescribed, as long it is unique. Consider best practices for determining IDs.

* Content type: `application/json`
* Status code: Choose an appropriate status code

Body (Exemplary message based on schema):
```
{
    "id": "OIdksUDwveiwe"
}
```

### Deletion of Webhook

Deletes a given webhook.

### - Request (DELETE)

```
Method: DELETE
Path: /dashboard/v1/notifications/{id}
```

* {id} is the ID returned during the webhook registration

**Response**

Implemented the response according to best practices.

### View *specific registered* webhook

Shows a specific webhook registration.

### - Request (GET)

```
Method: GET
Path: /dashboard/v1/notifications/{id}
```
* `{id}` is the ID for the webhook registration

**Response**

The response is similar to the POST request body, but further includes the ID assigned by the server upon adding the webhook.

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

Lists all registered webhooks.

### - Request (GET)

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

When a webhook is triggered, it should send information as follows. Where multiple webhooks are triggered, the information should be sent separately (i.e., one notification per triggered webhook).

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


## Endpoint 'Status': Monitoring service availability

The status interface indicates the availability of all individual services this service depends on. The reporting occurs based on status codes returned by the dependent services. The status interface further provides information about the number of registered webhooks (specification below), and the uptime of the service.

### - Request (GET)

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

# Additional requirements

* All endpoints should be *tested using automated testing facilities provided by Go (unit tests, httptest package)*. 
  * This includes the stubbing of the third-party endpoints to ensure test reliability (removing dependency on external services).
  * Include the testing of handlers using the httptest package. Your code should be structured to support this. 
  * Try to maximize test coverage as reported by Golang.
* Think about which information you can cache to minimise invocation on the third-party libraries. Use Firebase for this purpose.


# Deployment

The service is to be deployed on an IaaS solution OpenStack using Docker. 
URL to deployed service: <url....>
