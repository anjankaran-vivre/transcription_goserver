module.exports = {
  apps: [
    {
      name: "Backend-t_server-5050",
      script: "backend/build/transcript.exe",
      env: { PORT: 5050 }
    },
    {
      name: "Frontend-t_server-5063",
      script: "npm",
      args: "run dev",
      cwd: "frontend"
    },
  ]
};