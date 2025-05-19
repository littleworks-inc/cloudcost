# Cloud Cost Estimator for Infrastructure-as-Code (IaC)

A CLI tool and SaaS web service that estimates infrastructure costs based on IaC files like Terraform, Pulumi, CloudFormation, Azure Bicep, and Ansible.

## Features

- Parse IaC files (Terraform, Pulumi, CloudFormation, Bicep, Ansible)
- Estimate monthly/yearly costs using live pricing data from AWS, Azure, and GCP APIs
- Compare cost differences between current and proposed infrastructure changes
- Integrate with CI/CD pipelines (GitHub Actions, GitLab, Jenkins)
- Support user-friendly output formats (text, JSON, CSV, HTML)

## Installation

```bash
# Using Go
go install github.com/littleworks-inc/cloudcost@latest

# Or download the binary from GitHub Releases