import "@wailsio/runtime";
import { Settings } from "luxon";
import React from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import "./style.css";

const container = document.getElementById("root");

const root = createRoot(container!);

Settings.defaultLocale = "vi";

root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
