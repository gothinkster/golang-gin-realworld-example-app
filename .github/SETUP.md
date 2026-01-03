# CI/CD Setup Instructions

## Codecov Integration

To enable code coverage reporting to Codecov:

1. Sign up or log in to [Codecov](https://codecov.io/)
2. Add this repository to your Codecov account
3. Generate a Codecov token for this repository
4. Add the token as a GitHub repository secret:
   - Go to repository Settings → Secrets and variables → Actions
   - Click "New repository secret"
   - Name: `CODECOV_TOKEN`
   - Value: Paste your Codecov token
   - Click "Add secret"

Once the secret is configured, the CI workflow will automatically upload coverage reports to Codecov after test runs on Go 1.23.

## Note

The CI workflow is designed to work even without the Codecov token. Tests and other checks will still run successfully; only the coverage upload may have reduced functionality if the token is not configured.
