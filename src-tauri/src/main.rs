// Prevents additional console window on Windows in release mode
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use std::sync::Arc;
use std::time::Duration;
use tauri::Manager;
use tauri_plugin_shell::process::CommandEvent;
use tauri_plugin_shell::ShellExt;
use tokio::sync::{Mutex, RwLock};
use tokio::time::sleep;

/// Application state shared across commands.
struct AppState {
    backend_port: RwLock<Option<u16>>,
    sidecar_child: Mutex<Option<tauri_plugin_shell::process::CommandChild>>,
}

#[derive(serde::Serialize)]
struct BackendInfo {
    port: u16,
    base_url: String,
}

/// Tauri command: query the backend port after sidecar has started.
#[tauri::command]
async fn get_backend_info(state: tauri::State<'_, Arc<AppState>>) -> Result<BackendInfo, String> {
    let port = *state.backend_port.read().await;
    match port {
        Some(p) => Ok(BackendInfo {
            port: p,
            base_url: format!("http://127.0.0.1:{}", p),
        }),
        None => Err("Backend not ready yet".into()),
    }
}

/// Spawn the Go sidecar, read BACKEND_PORT from stdout, and wait for /health.
async fn start_backend_sidecar(app: &tauri::AppHandle) -> Result<u16, String> {
    let sidecar = app
        .shell()
        .sidecar("paperbanana-server")
        .map_err(|e| format!("sidecar config error: {}", e))?;

    let (mut rx, child) = sidecar
        .spawn()
        .map_err(|e| format!("failed to spawn sidecar: {}", e))?;

    // Store child handle for later cleanup
    {
        let state = app.state::<Arc<AppState>>();
        let mut guard = state.sidecar_child.lock().await;
        *guard = Some(child);
    }

    // Read stdout until we get BACKEND_PORT=xxxx
    let port = loop {
        match rx.recv().await {
            Some(CommandEvent::Stdout(line)) => {
                let output = String::from_utf8_lossy(&line);
                let trimmed = output.trim();
                if let Some(port_str) = trimmed.strip_prefix("BACKEND_PORT=") {
                    match port_str.parse::<u16>() {
                        Ok(p) => break p,
                        Err(_) => {
                            return Err(format!("invalid BACKEND_PORT value: {}", port_str));
                        }
                    }
                }
            }
            Some(CommandEvent::Stderr(line)) => {
                let output = String::from_utf8_lossy(&line);
                eprintln!("[sidecar stderr] {}", output.trim());
            }
            Some(CommandEvent::Terminated(status)) => {
                return Err(format!("sidecar exited early with status: {:?}", status));
            }
            Some(_) => continue,
            None => {
                return Err("sidecar stdout channel closed".into());
            }
        }
    };

    // Wait for health endpoint to confirm backend is serving
    wait_for_health(port, 30).await?;

    // Store port in state
    {
        let state = app.state::<Arc<AppState>>();
        let mut guard = state.backend_port.write().await;
        *guard = Some(port);
    }

    Ok(port)
}

/// Poll the backend health endpoint with retries.
async fn wait_for_health(port: u16, max_retries: u32) -> Result<(), String> {
    let client = reqwest::Client::builder()
        .timeout(Duration::from_secs(2))
        .build()
        .map_err(|e| e.to_string())?;

    for i in 0..max_retries {
        match client
            .get(format!("http://127.0.0.1:{}/health", port))
            .send()
            .await
        {
            Ok(resp) if resp.status().is_success() => return Ok(()),
            _ => {
                sleep(Duration::from_millis(300)).await;
            }
        }
    }

    Err(format!(
        "backend health check failed after {} retries on port {}",
        max_retries, port
    ))
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    let state = Arc::new(AppState {
        backend_port: RwLock::new(None),
        sidecar_child: Mutex::new(None),
    });

    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_single_instance::init(|app, _args, _cwd| {
            let _ = app.get_webview_window("main").map(|w| w.set_focus());
        }))
        .plugin(tauri_plugin_updater::Builder::new().build())
        .manage(state.clone())
        .invoke_handler(tauri::generate_handler![get_backend_info])
        .setup(move |app| {
            let handle = app.handle().clone();

            // Spawn sidecar in a background task
            tauri::async_runtime::spawn(async move {
                match start_backend_sidecar(&handle).await {
                    Ok(port) => {
                        println!("[tauri] backend ready on port {}", port);
                        // Show the main window once backend is ready
                        if let Some(window) = handle.get_webview_window("main") {
                            let _ = window.show();
                            let _ = window.set_focus();
                        }
                    }
                    Err(e) => {
                        eprintln!("[tauri] failed to start backend: {}", e);
                        // TODO: show error dialog to user
                    }
                }
            });

            Ok(())
        })
        .on_window_event(move |window, event| {
            if let tauri::WindowEvent::CloseRequested { .. } = event {
                let handle = window.app_handle().clone();
                tauri::async_runtime::spawn(async move {
                    let state = handle.state::<Arc<AppState>>();
                    let mut guard = state.sidecar_child.lock().await;
                    if let Some(child) = guard.take() {
                        let _ = child.kill();
                    }
                });
            }
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

fn main() {
    run();
}
