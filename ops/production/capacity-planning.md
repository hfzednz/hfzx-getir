# Capacity planning

## Inputs

- Orders/min forecast by city
- k6 results from `load` overlay
- HPA / KEDA headroom (CPU 65% prod target)
- GPU queue depth for AI

## Outputs

- Terraform node_max adjustments
- Helm replica floors for bff-customer, order, payment, realtime
- Pre-warm flags before marketing spikes

Review cadence: monthly + before each regional GA.
