# Antivirus API

~~Publicly hosted on~~

~~https://api.antivirushk.com/~~

Was taken down to reduce cost
## API List

### /getQuaratineBuildingCount

Get count for Compulsory Quarantine buildings


Example output
```
{ "count": 5874}
```

### /getQuaratineBuildingList

Get List of Compulsory Quarantine buildings

Query Parameters:

| Name | Description | Example value
| --- | --- | --- |
| start | Start index of record to be return | 1 (default=1) |
| count | Max. count of record to be returned | 200 (default=200) |
| district | Filter by district (value in Chinese) | 深水埗 |

Example call

```
https://api.antivirushk.com/getQuaratineBuildingList?count=1000&district=深水埗
```

```
https://api.antivirushk.com/getQuaratineBuildingList?start=1&count=300
```

Example output

```
{
    "data": [
        {
            "id": 839,
            "chiAddr": "ONE NEW YORK",
            "engAddr": "ONE NEW YORK",
            "district": "深水埗",
            "endDate": "2020-02-27",
            "lat": 22.332094,
            "lng": 114.14691
        }
    ],
    "count": 1
}
```

## Code

This piece of code is develop from
https://github.com/GoogleCloudPlatform/golang-samples/tree/master/cloudsql/mysql/database-sql

### Deploy to App Engine

Follow the instruction from the repo above

### Build locally

Install the following 2 packages

```
go get github.com/go-sql-driver/mysql
go get github.com/rs/cors
```

This project use Google Cloud SQL as the DB

Set the following environment variables to connect to a Cloud SQL instance

You may need to change the code if you are using a Mysql hosted somewhere else

```
# Replace INSTANCE_CONNECTION_NAME with the value obtained when configuring your
# Cloud SQL instance, available from the Google Cloud Console or from the Cloud SDK.
# For Cloud SQL 2nd generation instances, this should be in the form of "project:region:instance".
CLOUDSQL_CONNECTION_NAME: <INSTANCE_CONNECTION_NAME>
CLOUDSQL_USER: <DB user>
CLOUDSQL_PASSWORD: <DB password>
```

Build
```
go build
```

Run
```
./go-api
```