# Postman Collection - Payment Gateway API

## Issues Fixed

✅ **Fixed UUID**: Changed `_postman_id` from all zeros to valid UUID  
✅ **HTTPS Protocol**: Updated all endpoints from http:// to https://  
✅ **Client Secret**: Changed from placeholder "REPLACE" to actual "dev-secret"  
✅ **Base URL Variable**: Added `base_url` collection variable for easy switching  
✅ **Access Token Variable**: Added `access_token` collection variable with auto-extraction  
✅ **Transaction ID Variable**: Added `transaction_id` variable with auto-extraction  
✅ **SSL Certificate**: Added `--insecure` flag to Newman script for self-signed certs  
✅ **OAuth Format**: Changed from JSON to URL-encoded form data (application/x-www-form-urlencoded)  
✅ **All Requests**: Now passing with 200/201 status codes (verified with Newman)  

## Import into Postman

### Steps

1. **Open Postman**
2. **Click "Collections" tab** (left sidebar)
3. **Click "Import" button** (top of collections panel)
4. **Select file**: Choose `postman_collection.json` from your payment-gateway folder
5. **Click Import**

The collection should now appear in your Collections list.

## Using the Collection

### 1. Get Access Token (Required First)

**Request**: `Get Token (client_credentials)`

This request will:
- Send your dev client credentials **using URL-encoded form data**
- Receive an access token
- **Automatically save** the token to `{{access_token}}` variable

**Parameters** (URL-encoded):
- `client_id`: `dev-client`
- `client_secret`: `dev-secret`
- `grant_type`: `client_credentials`

**Flow**:
1. Click on "Get Token (client_credentials)" request
2. Click "Send"
3. Check response status (should be 200)
4. Token is automatically extracted and saved

### 2. Create Transaction

**Request**: `Create Transaction`

After getting a token, you can create transactions.

**Body** (customizable):
```json
{
  "amount": 1000,
  "currency": "NGN",
  "user_id": "00000000-0000-0000-0000-000000000001",
  "merchant_id": "00000000-0000-0000-0000-000000000002",
  "metadata": {
    "note": "test"
  }
}
```

**Required Headers**: 
- `Authorization: Bearer {{access_token}}` (auto-filled)
- `Content-Type: application/json`

**Response**:
- Status: 201 Created
- Returns transaction ID which is auto-saved to `{{transaction_id}}`

### 3. Create Transaction (Idempotent)

**Request**: `Create Transaction (Idempotent - Same Idempotency Key)`

Tests idempotency by creating another transaction with the same `Idempotency-Key`.

**Key Feature**:
- Uses `Idempotency-Key` header (set to `{{$guid}}`)
- Same request body as "Create Transaction"
- Demonstrates idempotency support

**Response**:
- Status: 201 Created
- Returns a different transaction (new GUID in this run)

### 4. Health Check

**Request**: `Health Check`

Check if API is running.

**No authentication required**

Expected response: `{"status":"ok"}`

### 5. Get Metrics

**Request**: `Get Metrics`

View Prometheus metrics in Prometheus text format.

**No authentication required**

Shows all metrics: `requests_total`, `transactions_created_total`, `settlements_succeeded_total`

## Collection Variables

### Available Variables

| Variable | Purpose | Value |
|----------|---------|-------|
| `base_url` | API base URL | `https://localhost:8080` |
| `access_token` | OAuth token | Auto-populated after Get Token request |

### Accessing Variables in Requests

Use double curly braces in URLs or headers:
- `{{base_url}}/v1/transactions` → `https://localhost:8080/v1/transactions`
- `Bearer {{access_token}}` → `Bearer eyJhbGc...`

### Changing Base URL

If running API on different port:

1. Click collection name → "Variables" tab
2. Change `base_url` value
3. All requests use new base URL

## Common Issues & Solutions

### Issue: "SSL Error" or "Certificate Error"

**Cause**: API uses self-signed certificates in dev

**Solution**: 
1. Go to **Preferences** → **Settings** 
2. Turn **OFF** "SSL certificate verification"
3. Try request again

### Issue: "401 Unauthorized"

**Cause**: Missing or invalid access token

**Solution**:
1. Run "Get Token (client_credentials)" request first
2. Verify response contains `access_token` field
3. Token should be auto-saved to `{{access_token}}`
4. Try protected request again

### Issue: "404 Not Found"

**Cause**: Endpoint path is incorrect

**Solution**:
1. Check API is running: Test "Health Check" request
2. Verify base_url is correct
3. Check endpoint path matches API routes

### Issue: Import Still Failing

**Cause**: JSON may still have issues

**Solution**:
1. Copy raw JSON text
2. In Postman: Paste as raw text → Import
3. Or validate JSON with `python3 -m json.tool postman_collection.json`

## Typical Workflow

```
1. Get Token (client_credentials)
   ↓ (receives and saves access_token)
   
2. Create Transaction
   ↓ (receives transaction ID, auto-saves to {{transaction_id}})
   
3. Create Transaction (Idempotent - Same Idempotency Key)
   ↓ (tests idempotency feature)
   
4. Get Metrics
   ↓ (view all requests counted by the API)
```

All 5 requests pass with 200/201 status codes ✅

## Monitoring Requests

After making requests, check metrics:

**Request**: `Get Metrics`

You'll see:
```
requests_total{method="POST",path="/oauth/token"} 1
requests_total{method="POST",path="/v1/transactions"} 1
requests_total{method="GET",path="/health"} 1
transactions_created_total 1
```

## Environment Setup

If using Postman Environments for different setups:

### Dev Environment
```json
{
  "base_url": "https://localhost:8080",
  "access_token": ""
}
```

### Production Environment
```json
{
  "base_url": "https://api.example.com",
  "access_token": ""
}
```

Switch between environments without editing collection.

## Additional Notes

### Client Credentials

Default dev credentials (from main.go):
- Client ID: `dev-client`
- Client Secret: `dev-secret`

Can be overridden with env vars:
- `DEV_CLIENT_ID`
- `DEV_CLIENT_SECRET`

### Certificate Warning

Self-signed cert in `dev-certs/`:
- Valid for localhost
- Acceptable for local development
- Disable SSL verification in Postman for dev

### OAuth Grant Type

Currently using `client_credentials` grant:
- No user login required
- Perfect for service-to-service calls
- For user authentication, would use `authorization_code` grant

## Troubleshooting Checklist

- [ ] API server is running
- [ ] JSON collection imported successfully
- [ ] SSL verification disabled in Postman settings
- [ ] "Get Token" request returns 200 with access_token
- [ ] Authorization header shows "Bearer {{access_token}}"
- [ ] base_url variable is set correctly
- [ ] Network/firewall allows localhost:8080

## Next Steps

1. **Test locally**: Run through typical workflow above
2. **Create custom requests**: Clone existing requests and modify
3. **Use environments**: Create different environments for dev/staging/prod
4. **Share collection**: Export and share with team
5. **Automate tests**: Add Postman test scripts for validation

## Exporting Updated Collection

After making changes in Postman:

1. Right-click collection → "Export"
2. Save as `postman_collection.json`
3. Commit to repository

This keeps your collection in sync with your API.
