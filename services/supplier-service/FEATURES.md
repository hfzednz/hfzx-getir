# Supplier Platform Features

- Supplier/partner onboarding, verification, documents, certifications, banking refs (opaque)
- Contracts with versioning, activate, renew
- RFQ → quotation → award → sourcing PO (ERP PO via port)
- Inbound ASN/shipment + QC receive; invoice match *signal* (ERP AP remains SoT)
- Marketplace sellers + listing refs (stock hints only)
- Catalog submissions → catalog-service port on approve
- EDI ingest (850/855/856/810/846)
- Scorecards, messaging, approvals, AI recommend/risk ports
- Portal snapshot + admin stats; outbox on `supplier.events`
