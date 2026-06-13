import React from "react";
import ReactDOM from "react-dom/client";
import { App } from "./App";
import { initAuth } from "./auth/keycloak";
import "./styles.css";

async function bootstrap() {
  await initAuth();

  ReactDOM.createRoot(document.getElementById("root")!).render(
    <React.StrictMode>
      <App />
    </React.StrictMode>
  );
}

void bootstrap();
