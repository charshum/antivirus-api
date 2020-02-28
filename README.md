# Antivirus API

Publicly hosted on

https://api.antivirushk.com/

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