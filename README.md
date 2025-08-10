# GOV.UK Street Manager relay

### References

-   https://department-for-transport-streetmanager.github.io/street-manager-docs/open-data/example-http-subscriber/
-   https://ip-ranges.amazonaws.com/ip-ranges.json
-   https://www.manage-roadworks.service.gov.uk/open-data-onboarding
-   JSON schema: https://department-for-transport-streetmanager.github.io/street-manager-docs/api-documentation/json/api-notification-event-notifier-message.json

### Misc Notes

-   `fgrep ARN-5210-27348242 *.json` shows 2 events CREATED/UPDATED for same object reference
    -   56165.706.json (Create)
    -   56210.476.json (Update)

## Attribution

-   Street Manager Open Data (GOV.UK, Department of Transport), https://department-for-transport-streetmanager.github.io/street-manager-docs/open-data/

## TODO & Future Enhancements

-   [ ] Improve README documentation
-   [ ] Add authentication and rate limiting
-   [ ] Support for additional spatial queries (e.g., radius search)
-   [ ] Pagination and filtering options
-   [ ] Docker Compose for easier setup
-   [ ] OpenAPI/Swagger documentation (auto-generated from code)
-   [ ] More robust error handling and logging
-   [ ] Unit and integration tests for import and API layers

## License

See `LICENSE.md` for further details.
