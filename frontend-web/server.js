'use strict';

const path = require('path');
const express = require('express');

const app = express();
const PORT = Number(process.env.PORT) || 5401;
const publicDir = path.join(__dirname, 'public');

app.disable('x-powered-by');
app.use(express.static(publicDir, { index: false }));

app.get('/', (_req, res) => {
  res.sendFile(path.join(publicDir, 'index.html'));
});

app.listen(PORT, '0.0.0.0', () => {
  // eslint-disable-next-line no-console
  console.log(`[xyKitchen web] http://0.0.0.0:${PORT}`);
});
