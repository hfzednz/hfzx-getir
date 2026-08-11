# Blue/Green cutover playbook

1. Deploy green revision (new tag) alongside blue.
2. Run `tools/prod-validate` against green ingress/shadow host.
3. Shift 1% mesh weight → watch 15m.
4. Shift 100% on GO.
5. Keep blue warm 24h; delete only after POST_RELEASE T+24h green.
