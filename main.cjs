const { app, BrowserWindow, Menu } = require("electron");

const createWindow = () => {
  Menu.setApplicationMenu(null);

  const win = new BrowserWindow({
    show: false
  });
  win.loadURL("http://localhost:7500");
  win.maximize();
  win.show();
};

app.whenReady().then(() => {
  createWindow();
});

app.on("window-all-closed", () => {
  app.quit();
});
