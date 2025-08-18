# GOV.UK Street Manager Relay

This project is a Go-based relay for the GOV.UK Street Manager API. It receives notifications from the Street Manager API via Amazon SNS, processes them, and stores them in a local SQLite database. It also provides a REST API to search for events stored in the database.

## Architecture

The project is built with a modular architecture, with different components responsible for specific functionalities.

-   **`main.go`**: The entry point of the application. It uses the `cobra` library to define the command-line interface for the application.
-   **`cmd/api_server.go`**: This file sets up the Gin-based HTTP server. It configures middleware for logging, metrics (Prometheus), compression, CORS, and health checks.
-   **`cmd/bulk_loader.go`**: This file contains the logic for bulk loading data from a folder into the SQLite database.
-   **`cmd/regen_rtree.go`**: This file contains the logic for regenerating the R-tree index in the database, which is used for spatial queries.
-   **`internal/db.go`**: This file handles all the database interactions. It uses the `sqlite3` library to work with the SQLite database.
-   **`internal/routes/sns.go`**: This file defines the handler for incoming SNS messages. It validates the message signature and then processes the message based on its type (`SubscriptionConfirmation` or `Notification`).
-   **`internal/routes/search.go`**: This file defines the handler for the `/v1/street-manager-relay/search` endpoint. It parses the bounding box and facet parameters from the query string and then uses the `DbRepository` to search for events in the database.
-   **`internal/routes/refdata.go`**: This file defines the handler for the `/v1/street-manager-relay/refdata` endpoint. It returns reference data used for filtering and faceting event searches.
-   **`models/*`**: These files define the data models used in the application, such as `Event`, `BoundingBox`, and `Facets`.

## Installation

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/rm-hull/street-manager-relay.git
    cd street-manager-relay
    ```

2.  **Install dependencies:**

    ```bash
    go mod tidy
    ```

3.  **Build the application:**

    ```bash
    go build
    ```

## Usage

### API Endpoints

#### `POST /v1/street-manager-relay/sns`

This endpoint is used to receive SNS messages from the GOV.UK Street Manager API. It handles `SubscriptionConfirmation` and `Notification` messages. You don't need to interact with this endpoint directly. It's designed to be used by the Amazon SNS service.

#### `GET /v1/street-manager-relay/search`

This endpoint is used to search for events in the database.

**Parameters:**

-   `bbox` (required): A comma-separated string of four coordinates representing the bounding box for the search (e.g., `min_easting,max_easting,min_northing,max_northing`).
-   **Facets** (optional): You can filter the search results by providing one or more of the following facet parameters. You can provide multiple values for each facet by either repeating the parameter (e.g., `work_status_ref=planned&work_status_ref=in_progress`) or by providing a comma-separated list of values (e.g., `work_status_ref=planned,in_progress`).

    -   `permit_status`
    -   `traffic_management_type_ref`
    -   `work_status_ref`
    -   `work_category_ref`
    -   `road_category`
    -   `highway_authority`
    -   `promoter_organisation`

**Example `curl` request:**

```bash
curl -X GET "http://localhost:8080/v1/street-manager-relay/search?bbox=418995,435778,429089,441777&work_status_ref=in_progress,planned"
```

#### `GET /v1/street-manager-relay/refdata`

This endpoint returns reference data used for filtering and faceting event searches. The data includes lists of possible values for facets such as permit status, traffic management type, work status, work category, road category, highway authority, and promoter organisation, along with counts for each value.

**Response:**

-   `refdata`: An object mapping each facet to its possible values and their counts.
-   `attribution`: Attribution information for the data source.

**Example response:**

```json
{
  "refdata": {
    "permit_status": { "granted": 123, "refused": 45 },
    "work_status_ref": { "planned": 67, "in_progress": 89 },
    ...
  },
  "attribution": "Contains public sector information licensed under the Open Government Licence v3.0."
}
```

**Example `curl` request:**

```bash
curl -X GET "http://localhost:8080/v1/street-manager-relay/refdata"
```

### Command-Line Interface

The application provides a command-line interface to manage the database.

-   **`api-server`**: Starts the HTTP API server.

    ```bash
    ./street-manager-relay api-server --port 8080
    ```

-   **`bulk-loader`**: Bulk loads data from a folder into the database.

    ```bash
    ./street-manager-relay bulk-loader <folder>
    ```

-   **`regen`**: Regenerates the R-tree index in the database.

    ```bash
    ./street-manager-relay regen
    ```

## Dependencies

-   [Gin](https://github.com/gin-gonic/gin): A popular web framework for Go.
-   [Cobra](https://github.com/spf13/cobra): A library for creating powerful modern CLI applications.
-   [SQLite3](https://github.com/mattn/go-sqlite3): A driver for SQLite.
-   [Gin-Prometheus](https://github.com/Depado/ginprom): A middleware for exporting Prometheus metrics.
-   [Go-Memoize](https://github.com/kofalt/go-memoize): A library for memoizing function calls.

## References

-   https://department-for-transport-streetmanager.github.io/street-manager-docs/open-data/example-http-subscriber/
-   https://ip-ranges.amazonaws.com/ip-ranges.json
-   https://www.manage-roadworks.service.gov.uk/open-data-onboarding
-   JSON schema: https://department-for-transport-streetmanager.github.io/street-manager-docs/api-documentation/json/api-notification-event-notifier-message.json

## Misc Notes

-   `fgrep ARN-5210-27348242 *.json` shows 2 events CREATED/UPDATED for same object reference
    -   56165.706.json (Create)
    -   56210.476.json (Update)

## TODO & Future Enhancements

-   [x] Improve README documentation
-   [ ] Add authentication and rate limiting
-   [ ] Support for additional spatial queries (e.g., radius search)
-   [ ] Pagination and filtering options
-   [ ] Docker Compose for easier setup
-   [ ] OpenAPI/Swagger documentation (auto-generated from code)
-   [ ] More robust error handling and logging
-   [ ] Unit and integration tests for import and API layers

## License

This project is licensed under the MIT License. See the `LICENSE.md` file for details.

## Attribution

-   Street Manager Open Data (GOV.UK, Department of Transport), https://department-for-transport-streetmanager.github.io/street-manager-docs/open-data/
