const https = require('https');
const fs = require('fs');
const path = require('path');

const certDir = path.join(__dirname, 'certs');
const keyPath = path.join(certDir, 'key.pem');
const certPath = path.join(certDir, 'cert.pem');

if (!fs.existsSync(certDir)) {
  fs.mkdirSync(certDir, { recursive: true });
}

const options = {
  key: fs.readFileSync(keyPath),
  cert: fs.readFileSync(certPath)
};

const server = https.createServer(options, (req, res) => {
  res.writeHead(200);
  res.end('HTTPS server is running!');
});

server.listen(3001, () => {
  console.log('HTTPS server running on https://localhost:3001');
});
