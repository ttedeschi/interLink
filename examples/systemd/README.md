# Install SLURM plugin as a systemd

```bash
VERSION=0.2.2

wget -O $HOME/.interlink/bin/slurm-plugin  https://github.com/interTwin-eu/interLink/releases/download/${VERSION}/interlink-sidecar-slurm_Linux_x86_64
chmod +x $HOME/.interlink/bin/slurm-plugin

mkdir -p $HOME/.config/systemd/user

cat <<EOF > $HOME/.config/systemd/user/slurm-plugin.service
[Unit]
Description=This Unit is needed to automatically start the SLURM sidecar at system startup
After=network.target

[Service]
Type=simple
ExecStart=$HOME/.interlink/bin/slurm-plugin
Environment="INTERLINKCONFIGPATH=$HOME/.interlink/config/InterLinkConfig.yaml"
StandardOutput=file:$HOME/.interlink/logs/plugin.log
StandardError=file:$HOME/.interlink/logs/plugin.log

[Install]
WantedBy=multi-user.target
EOF

systemctl --user daemon-reload
systemctl --user enable slurm-plugin.service

systemctl --user start slurm-plugin.service
```
