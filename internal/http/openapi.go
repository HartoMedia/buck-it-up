package http

// ...existing code...

// GetOpenAPISpec returns a minimal OpenAPI 3.0 JSON document describing the public API.
// This spec is generated to match the endpoints registered in router.go so you can load it
// in Swagger UI at /swagger (which loads /openapi.json).
func GetOpenAPISpec() []byte {
	return []byte(`{
  "openapi": "3.0.0",
  "info": {
    "title": "Buck It Up API",
    "version": "1.0.0",
    "description": "OpenAPI spec for Buck It Up with Bearer Auth and custom LIST method described using vendor extensions."
  },
  "servers": [
    { "url": "http://localhost:8080" }
  ],

  "components": {
    "securitySchemes": {
      "bearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT"
      }
    },
    "schemas": {
      "NewBucket": {
        "type": "object",
        "properties": {
          "name": { "type": "string" }
        },
        "required": ["name"]
      },
      "RecreateAccessKeyRequest": {
        "type": "object",
        "properties": {
          "role": {
            "type": "string",
            "enum": ["readOnly", "readWrite", "all"]
          }
        },
        "required": ["role"]
      },
      "UploadObjectRequest": {
        "type": "object",
        "properties": {
          "object_key": { "type": "string" },
          "content": { "type": "string" },
          "content_type": { "type": "string" },
          "base64_encoded": { "type": "boolean" }
        },
        "required": ["object_key", "content"]
      }
    }
  },

  "security": [
    { "bearerAuth": [] }
  ],

  "paths": {
    "/health": {
      "get": {
        "security": [],
        "summary": "Health check",
        "responses": {
          "200": { "description": "ok" }
        }
      }
    },

    "/echo": {
      "get": {
        "security": [],
        "summary": "Echo request body",
        "description": "Echoes the request body back.",
        "requestBody": {
          "content": {
            "application/octet-stream": {}
          },
          "required": false
        },
        "responses": {
          "200": { "description": "echoed body" }
        }
      }
    },

    "/": {
      "get": {
        "summary": "List buckets (GET fallback)",
        "description": "Primary method is LIST (custom HTTP verb), but GET is provided for tooling compatibility.",
        "responses": {
          "200": { "description": "array of buckets" }
        }
      },
      "x-list": {
        "summary": "List buckets (custom HTTP LIST method)",
        "description": "This endpoint supports a custom HTTP method LIST, which is the real method used by the router.",
        "responses": {
          "200": { "description": "array of buckets" }
        }
      },
      "post": {
        "summary": "Create bucket",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/NewBucket" }
            }
          }
        },
        "responses": {
          "201": { "description": "bucket created" },
          "409": { "description": "bucket exists" }
        }
      }
    },

    "/{name}": {
      "get": {
        "summary": "Get bucket by name",
        "parameters": [
          { "name": "name", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": { "description": "bucket" },
          "404": { "description": "not found" }
        }
      },
      "delete": {
        "summary": "Delete bucket by name",
        "parameters": [
          { "name": "name", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "204": { "description": "deleted" },
          "409": { "description": "bucket not empty" }
        }
      }
    },

    "/{name}/access-keys": {
      "get": {
        "summary": "List access keys for a bucket",
        "parameters": [
          { "name": "name", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": { "description": "array of access keys" }
        }
      }
    },

    "/{name}/access-keys/recreate": {
      "post": {
        "summary": "Recreate an access key for a bucket",
        "parameters": [
          { "name": "name", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/RecreateAccessKeyRequest" }
            }
          }
        },
        "responses": {
          "201": { "description": "new access key" }
        }
      }
    },

    "/{bucketName}": {
      "get": {
        "summary": "List objects in a bucket",
        "parameters": [
          { "name": "bucketName", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": { "description": "array of objects" }
        }
      }
    },

    "/{bucketName}/upload": {
      "post": {
        "summary": "Upload an object to a bucket",
        "parameters": [
          { "name": "bucketName", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/UploadObjectRequest" }
            }
          }
        },
        "responses": {
          "201": { "description": "object created" },
          "409": { "description": "object already exists" }
        }
      }
    },

    "/{bucketName}/all/{objectKey}": {
      "get": {
        "summary": "Get object (metadata + base64 content)",
        "parameters": [
          { "name": "bucketName", "in": "path", "required": true, "schema": { "type": "string" } },
          { "name": "objectKey", "in": "path", "required": true, "schema": { "type": "string" }, "description": "Object key; may include slashes" }
        ],
        "responses": {
          "200": { "description": "object with base64 content" },
          "404": { "description": "not found" }
        }
      }
    },

    "/{bucketName}/metadata/{objectKey}": {
      "get": {
        "summary": "Get object metadata only",
        "parameters": [
          { "name": "bucketName", "in": "path", "required": true, "schema": { "type": "string" } },
          { "name": "objectKey", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": { "description": "object metadata" },
          "404": { "description": "not found" }
        }
      }
    },

    "/{bucketName}/content/{objectKey}": {
      "get": {
        "summary": "Get object raw content (streamed)",
        "parameters": [
          { "name": "bucketName", "in": "path", "required": true, "schema": { "type": "string" } },
          { "name": "objectKey", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": { "description": "raw content" },
          "404": { "description": "not found" }
        }
      }
    },

    "/{bucketName}/{objectKey}": {
      "delete": {
        "summary": "Delete an object",
        "parameters": [
          { "name": "bucketName", "in": "path", "required": true, "schema": { "type": "string" } },
          { "name": "objectKey", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "204": { "description": "deleted" },
          "404": { "description": "not found" }
        }
      }
    }
  }
}

`)
}

// GetSwaggerHTML returns a small HTML page that loads Swagger UI from CDN and points to /openapi.json
func GetSwaggerHTML() []byte {
	return []byte(`<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>Buck It Up - Swagger UI</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@4/swagger-ui.css" />
    <style>body { margin:0; padding:0; }</style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@4/swagger-ui-bundle.js"></script>
    <script>
      window.onload = function() {
        const ui = SwaggerUIBundle({
          url: '/openapi.json',
          dom_id: '#swagger-ui',
          presets: [SwaggerUIBundle.presets.apis],
          layout: 'BaseLayout'
        })
      }
    </script>
  </body>
</html>`)
}
