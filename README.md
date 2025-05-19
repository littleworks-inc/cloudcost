Great! Here's the final version of the README with the Current Implementation Status and Roadmap sections added, but without the Getting Help section:

```markdown
# Cloud Cost Estimator for Infrastructure-as-Code (IaC)

A CLI tool and SaaS web service that estimates infrastructure costs based on IaC files like Terraform, Pulumi, CloudFormation, Azure Bicep, and Ansible.

## Features

- Parse IaC files (Terraform, Pulumi, CloudFormation, Bicep, Ansible)
- Estimate monthly/yearly costs using live pricing data from AWS, Azure, and GCP APIs
- Compare cost differences between current and proposed infrastructure changes
- Integrate with CI/CD pipelines (GitHub Actions, GitLab, Jenkins)
- Support user-friendly output formats (text, JSON, CSV, HTML)

## Current Implementation Status

The project is under active development. Currently implemented features:

- ✅ Terraform parser for extracting resources
- ✅ AWS pricing client for retrieving real-time pricing data
- ✅ Command-line interface with estimate and diff commands
- ✅ Text, CSV, and HTML output formatters

Coming soon:
- 🔜 Ansible parser
- 🔜 CloudFormation parser
- 🔜 Azure pricing client
- 🔜 GCP pricing client
- 🔜 Web dashboard for visual analysis

## Installation

### Using Go

```bash
go install github.com/littleworks-inc/cloudcost@latest
```

### Binary Releases

You can download pre-built binaries from the [GitHub Releases](https://github.com/littleworks-inc/cloudcost/releases) page.

For macOS:
```bash
curl -L https://github.com/littleworks-inc/cloudcost/releases/latest/download/cloudcost-darwin-amd64 -o cloudcost
chmod +x cloudcost
```

For Linux:
```bash
curl -L https://github.com/littleworks-inc/cloudcost/releases/latest/download/cloudcost-linux-amd64 -o cloudcost
chmod +x cloudcost
```

For Windows (PowerShell):
```powershell
Invoke-WebRequest -Uri https://github.com/littleworks-inc/cloudcost/releases/latest/download/cloudcost-windows-amd64.exe -OutFile cloudcost.exe
```

## Quick Start

### Estimate cloud costs from IaC files

```bash
# Estimate costs from Terraform files
cloudcost estimate --path ./terraform-project

# Estimate costs from Ansible playbooks
cloudcost estimate --path ./ansible-playbooks

# Specify output format
cloudcost estimate --path ./terraform-project --output json

# Save output to a file
cloudcost estimate --path ./terraform-project --output-file cost-report.json
```

### Compare costs between versions

```bash
# Compare current IaC with a previous cost report
cloudcost diff --path ./terraform-project --compare-to previous-report.json
```

## Commands

### `estimate`

Estimate cloud costs from Infrastructure-as-Code files.

```bash
cloudcost estimate --path PATH [flags]
```

**Flags:**
- `--path string` - Path to IaC files (required)
- `--output-file string` - File to save the report to
- `--output string` - Output format (text, json, csv, html) (default "text")
- `--config string` - Config file (default is $HOME/.cloudcost.yaml)

### `diff`

Compare costs between IaC versions.

```bash
cloudcost diff --path PATH --compare-to REPORT_FILE [flags]
```

**Flags:**
- `--path string` - Path to IaC files (required)
- `--compare-to string` - Previous cost report to compare against (required)
- `--output string` - Output format (text, json, csv, html) (default "text")
- `--config string` - Config file (default is $HOME/.cloudcost.yaml)

### `version`

Display the version, commit, and build date of the tool.

```bash
cloudcost version
```

## Configuration

You can configure the Cloud Cost Estimator using a configuration file or environment variables.

### Configuration File

The default configuration file is located at `$HOME/.cloudcost.yaml`. You can specify a different configuration file using the `--config` flag.

Example configuration file:

```yaml
# General settings
general:
  currency: USD
  output_format: text
  output_file: ""

# Cloud provider settings
providers:
  aws:
    enabled: true
    regions:
      - us-east-1
      - us-west-2
      - eu-west-1
  
  azure:
    enabled: true
    regions:
      - eastus
      - westus2
      - westeurope
  
  gcp:
    enabled: true
    regions:
      - us-central1
      - us-east4
      - europe-west1

# IaC parser settings
parsers:
  terraform:
    enabled: true
    plan_file: ""
  
  pulumi:
    enabled: true
    preview_file: ""
  
  cloudformation:
    enabled: true
    template_file: ""
  
  azure_arm:
    enabled: true
    template_file: ""
  
  ansible:
    enabled: true
    playbook_file: ""

# Pricing settings
pricing:
  cache_ttl: 3600
  cache_dir: ""
  reserved_instances: false
  savings_plans: false
  spot_instances: false

# Resource filter settings
filters:
  include_types: []
  exclude_types: []
  include_tags: {}
  exclude_tags: {}
```

### Environment Variables

You can also configure the Cloud Cost Estimator using environment variables:

- `CLOUDCOST_CURRENCY` - Currency to use for pricing (default: USD)
- `CLOUDCOST_OUTPUT_FORMAT` - Default output format (default: text)
- `CLOUDCOST_CACHE_TTL` - Cache TTL in seconds (default: 3600)
- `CLOUDCOST_CACHE_DIR` - Directory to store cache files
- `AWS_ACCESS_KEY_ID` - AWS access key ID for accessing the Pricing API
- `AWS_SECRET_ACCESS_KEY` - AWS secret access key for accessing the Pricing API
- `AWS_REGION` - AWS region to use for credentials (default: us-east-1)
- `AZURE_CLIENT_ID` - Azure client ID for accessing the Azure Retail Prices API
- `AZURE_CLIENT_SECRET` - Azure client secret for accessing the Azure Retail Prices API
- `AZURE_TENANT_ID` - Azure tenant ID for accessing the Azure Retail Prices API
- `GOOGLE_APPLICATION_CREDENTIALS` - Path to GCP service account key file

## Supported IaC Formats

- **Terraform**: HCL files (.tf) and plan files (.tfplan)
- **Pulumi**: Preview JSON output from `pulumi preview --json`
- **CloudFormation**: Template files (.yaml, .json, .template)
- **Azure ARM/Bicep**: Template files (.json, .bicep)
- **Ansible**: Playbook files (.yml, .yaml)

## Supported Cloud Providers

- **AWS**: EC2, RDS, Lambda, S3, ECS, EKS, and more
- **Azure**: Virtual Machines, App Service, SQL Database, and more
- **GCP**: Compute Engine, Cloud SQL, Cloud Functions, and more

## AWS Credentials Setup

The Cloud Cost Estimator requires AWS credentials to access the AWS Pricing API. You can configure credentials in several ways:

1. **AWS CLI** (recommended):
   ```bash
   aws configure
   ```
   This will prompt you for your AWS access key, secret key, and default region.

2. **Environment variables**:
   ```bash
   export AWS_ACCESS_KEY_ID=your_access_key
   export AWS_SECRET_ACCESS_KEY=your_secret_key
   export AWS_REGION=us-east-1
   ```

3. **Credentials file**:
   Create a file at `~/.aws/credentials` with the following content:
   ```
   [default]
   aws_access_key_id = your_access_key
   aws_secret_access_key = your_secret_key
   ```

   And a file at `~/.aws/config` with:
   ```
   [default]
   region = us-east-1
   ```

4. **EC2 Instance Role**:
   If running on EC2, attach an IAM role with the appropriate permissions.

The minimum IAM permissions required are:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "pricing:GetProducts",
                "pricing:DescribeServices"
            ],
            "Resource": "*"
        }
    ]
}
```

## Roadmap

- Complete additional IaC parsers and cloud pricing clients
- Add support for reserved instances and savings plans
- Release web dashboard and API
- Add AI-powered cost optimization suggestions

## CI/CD Integration

### GitHub Actions

```yaml
name: Estimate Cloud Costs

on:
  pull_request:
    paths:
      - 'terraform/**'
      - 'cloudformation/**'
      - 'ansible/**'

jobs:
  estimate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install Cloud Cost Estimator
        run: |
          curl -L https://github.com/littleworks-inc/cloudcost/releases/latest/download/cloudcost-linux-amd64 -o cloudcost
          chmod +x cloudcost
      
      - name: Estimate costs
        run: |
          ./cloudcost estimate --path ./terraform --output json --output-file cost-report.json
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          AWS_REGION: us-east-1
      
      - name: Post cost summary as comment
        uses: actions/github-script@v6
        with:
          script: |
            const fs = require('fs');
            const report = JSON.parse(fs.readFileSync('cost-report.json', 'utf8'));
            
            const summary = `## Cloud Cost Estimate
            
            - Monthly Cost: $${report.total_monthly.toFixed(2)}
            - Yearly Cost: $${report.total_yearly.toFixed(2)}
            - Resources: ${report.resources.length}
            
            [View detailed report](${process.env.GITHUB_SERVER_URL}/${process.env.GITHUB_REPOSITORY}/actions/runs/${process.env.GITHUB_RUN_ID})
            `;
            
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: summary
            });
```

## Building from Source

```bash
# Clone the repository
git clone https://github.com/littleworks-inc/cloudcost.git
cd cloudcost

# Build the binary
make build

# Run tests
make test

# Install the binary
make install
```

## Troubleshooting

### All resources show $0.00 prices

If all resources show $0.00 prices, this may be due to one of the following reasons:

1. **Missing AWS credentials**: Make sure you have configured AWS credentials as described in the "AWS Credentials Setup" section.
2. **AWS Pricing API access issues**: Ensure your AWS credentials have the appropriate permissions to access the AWS Pricing API.
3. **Unsupported resource types**: Some resource types may not be supported for pricing yet.

### Error: "failed to get pricing data"

This error indicates that the AWS Pricing API call failed. Possible reasons include:

1. **Invalid AWS credentials**: Check that your AWS credentials are correct.
2. **Network connectivity issues**: Ensure your system can connect to AWS services.
3. **AWS service outage**: Check the AWS Service Health Dashboard for any reported outages.

## Contributing

We welcome contributions from the community! Please feel free to submit pull requests, create issues, or suggest new features.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -am 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the [MIT License](LICENSE).
```