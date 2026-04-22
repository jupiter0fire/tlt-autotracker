const http = require('http');

const options = {
  host: '127.0.0.1',
  port: 17027,
  headers: {
    'Connection': 'Upgrade',
    'Upgrade': 'websocket',
    'Sec-WebSocket-Key': 'dGhlIHNhbXBsZSBub25jZQ==',
    'Sec-WebSocket-Version': '13'
  }
};

const req = http.request(options);

req.on('upgrade', (res, socket, upgradeHead) => {
  // Simple frame construction for {"type":"handshake"}
  // 0x81 (Fin + Text frame)
  // Mask bit set (0x80) + length 20 (0x14)
  const handshake = '{"type":"handshake"}';
  const mask = Buffer.from([0x01, 0x02, 0x03, 0x04]);
  const maskedPayload = Buffer.alloc(handshake.length);
  for (let i = 0; i < handshake.length; i++) {
    maskedPayload[i] = handshake.charCodeAt(i) ^ mask[i % 4];
  }
  const frame = Buffer.concat([Buffer.from([0x81, 0x80 | handshake.length]), mask, maskedPayload]);
  socket.write(frame);

  let buffer = Buffer.alloc(0);
  socket.on('data', (chunk) => {
    buffer = Buffer.concat([buffer, chunk]);
    while (buffer.length >= 2) {
      const secondByte = buffer[1];
      const length = secondByte & 0x7F;
      let payloadOffset = 2;
      let actualLength = length;
      if (length === 126) {
        if (buffer.length < 4) break;
        actualLength = buffer.readUInt16BE(2);
        payloadOffset = 4;
      } else if (length === 127) {
        if (buffer.length < 10) break;
        // Simplified skip for large payloads
        payloadOffset = 10;
      }
      
      if (buffer.length < payloadOffset + actualLength) break;
      
      const payload = buffer.slice(payloadOffset, payloadOffset + actualLength);
      // Opcode 1 is text
      if ((buffer[0] & 0x01) === 0x01) {
          console.log(payload.toString());
      }
      
      buffer = buffer.slice(payloadOffset + actualLength);
    }
  });
});

req.end();
