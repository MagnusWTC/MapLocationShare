const selfsigned = require('selfsigned');
const fs = require('fs');
const path = require('path');

const certDir = path.join(__dirname, 'certs');
if (!fs.existsSync(certDir)) {
  fs.mkdirSync(certDir, { recursive: true });
}

const attrs = [{ name: 'commonName', value: 'localhost' }];
const opts = {
  keySize: 2048,
  days: 365,
  algorithm: 'sha256',
  extensions: [
    {
      name: 'basicConstraints',
      cA: true,
    },
    {
      name: 'keyUsage',
      keyCertSign: true,
      digitalSignature: true,
      nonRepudiation: true,
      keyEncipherment: true,
      dataEncipherment: true,
    },
    {
      name: 'extKeyUsage',
      serverAuth: true,
      clientAuth: true,
      codeSigning: true,
      timeStamping: true,
    },
    {
      name: 'subjectAltName',
      altNames: [
        { type: 2, value: 'localhost' },
        { type: 2, value: '192.168.2.126' },
        { type: 7, ip: '127.0.0.1' },
        { type: 7, ip: '192.168.2.126' },
      ],
    },
  ],
};

const pems = selfsigned.generate(attrs, opts);

fs.writeFileSync(path.join(certDir, 'key.pem'), pems.private);
fs.writeFileSync(path.join(certDir, 'cert.pem'), pems.cert);

console.log('证书已生成！');
console.log('证书路径:', certDir);
