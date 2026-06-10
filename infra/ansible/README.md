# Ubuntu production server provisioning

This Ansible stack provisions an Ubuntu 22.04+ host with a practical production baseline:

- Non-root sudo admin user with SSH keys.
- SSH hardening with password login disabled by default.
- UFW firewall with default-deny inbound policy, fail2ban, unattended security updates, chrony time sync, logrotate, journald retention, and basic sysctl hardening.
- Docker Engine, the Docker Compose plugin, and a `docker-compose` compatibility command.
- Optional compose-managed Forgejo Actions runner for Codeberg.
- Optional compose-managed Woodpecker agent for Codeberg-hosted Woodpecker.

## Layout

```text
infra/ansible/
  ansible.cfg
  inventory/prod.example.yml
  group_vars/prod_servers.yml
  playbooks/prod-server.yml
  roles/
```

## First run

Install Ansible on your workstation, then create a local inventory from the example. The playbook uses Ansible core modules only.

```sh
cd infra/ansible
cp inventory/prod.example.yml inventory/prod.local.yml
```

Edit `inventory/prod.local.yml` with the server IP and bootstrap SSH user. Then set at least `admin_ssh_public_keys` in `group_vars/prod_servers.yml` or in an ignored local vars file.

Run the playbook:

```sh
ansible-playbook -i inventory/prod.local.yml playbooks/prod-server.yml
```

After the first run, connect as the configured admin user:

```sh
ssh deploy@SERVER_IP
```

## Important variables

| Variable | Default | Purpose |
| --- | --- | --- |
| `admin_user` | `deploy` | Non-root sudo user to create. |
| `admin_ssh_public_keys` | placeholder | Public keys installed for `admin_user`. |
| `ssh_password_authentication` | `false` | Disables password SSH when keys are configured. |
| `ufw_allowed_tcp_ports` | `22`, `80`, `443` | TCP ports opened in UFW. |
| `swap_file_size_mb` | `0` | Set to a positive number to create a swap file. |
| `docker_users` | `deploy` | Users added to the `docker` group. |

## Forgejo Actions runner

Enable the optional runner:

```yaml
forgejo_runner_enabled: true
forgejo_runner_registered: false
forgejo_runner_name: codeberg-prod-1
```

Run the playbook. It will start the runner container in a waiting state so you can register it without storing a token in git:

```sh
ssh deploy@SERVER_IP
cd /opt/forgejo-runner
sudo docker compose exec runner /bin/sh
forgejo-runner register --no-interactive \
  --instance https://codeberg.org/ \
  --name codeberg-prod-1 \
  --token RUNNER_TOKEN \
  --labels "docker:docker://node:22-bookworm,ubuntu-24.04:docker://ubuntu:24.04"
exit
```

Then set `forgejo_runner_registered: true` and rerun the playbook to start the runner daemon.

## Woodpecker agent

Enable the optional Codeberg-hosted Woodpecker agent:

```yaml
woodpecker_agent_enabled: true
woodpecker_agent_secret: "use-ansible-vault-or-an-ignored-vars-file"
woodpecker_agent_labels:
  location: prod-1
```

The Woodpecker agent uses `/var/run/docker.sock`. Only use this on hosts where the pipelines are trusted.

## Notes

- Keep real secrets in Ansible Vault or ignored local vars files.
- Running CI jobs executes remote code. Use dedicated hosts or VMs per trust boundary.
- If you change `ssh_port`, update `ufw_allowed_tcp_ports` at the same time.
