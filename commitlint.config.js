const Config = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    "scope-enum": [
      2,
      "always",
      [
        "auth",
        "users",
        "media",
        "chats",
        "notifications",
        "messages",
        "posts",
        "infra",
      ]
    ]
  }
};

export default Config;
