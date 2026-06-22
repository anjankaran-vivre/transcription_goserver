module.exports = {
  apps: [
    {
      name: "Backend-t_server-5050",
      script: "backend/server",
      env: { PORT: 5050 },

      // log files
      out_file: "/home/ubuntu/Vivre_Projects/transcription-server/backend/logs/out.log",
      error_file: "/home/ubuntu/Vivre_Projects/transcription-server/backend/logs/error.log",
      log_date_format: "YYYY-MM-DD HH:mm:ss",

      autorestart: true
    },
    {
      name: "Frontend-t_server-5063",
      script: "npm",
      args: "run dev",
      cwd: "frontend"
    },
  ]
};
