# How to Add Custom Breaking-Changes Checks

## Unit Test
1. Add a unit test for your scenario in one of the test files under [checker](../checker) with a comment "BC: \<use-case\> is breaking"
2. Add any accompanying OpenAPI specs under [data](../data)

## Localized Messages
1. Add localized texts under [checker/localizations_src](../checker/localizations_src)
2. Update [localization source file](../checker/localizations/localizations.go):
    ```
    go-localize -input checker/localizations_src -output checker/localizations
    ```   
    To install go-localize:
    ```
    go install github.com/m1/go-localize@latest
    ```
3. Make sure that [checker/localizations/localizations.go](../checker/localizations/localizations.go) contains the new messages

## Write the Checker Function
1. Create new go file under [checker](../checker) and name it by the breaking change use case
2. Create a check func inside the file and name it accordingly
3. Add the checker func to the defaultChecks or optionalChecks list

## Documentation
Optionally, add additional unit tests and comment them with "BC: \<use-case\> is breaking" or "BC: \<use-case\> is not breaking"

## Example
See this example of adding a custom check: https://github.com/oasdiff/oasdiff/pull/208/files
