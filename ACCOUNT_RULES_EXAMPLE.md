# Account Rules Feature

This implementation adds regex-based account prediction rules to Paisa's import system.

## Overview

The new feature provides:

1. **Account Rules Management UI**: Located at `/more/account-rules`
2. **New Template Helper**: `predictAccountWithRules` function for import templates
3. **API Endpoints**: For managing account rules in the configuration

## How to Use

### 1. Create Account Rules

1. Navigate to **More > Account Rules** in the Paisa interface
2. Click **"New Rule"** to create a rule
3. Fill in the rule details:
   - **Name**: A descriptive name (e.g., "Amazon Purchases")
   - **Pattern**: A regex pattern (e.g., `AMAZON.*|AMZ.*`)
   - **Account**: Target account (e.g., `Expenses:Shopping:Online`)
   - **Description**: Optional description
   - **Enabled**: Toggle to enable/disable the rule

### 2. Update Import Templates

Replace `predictAccount` with `predictAccountWithRules` in your import templates:

**Before:**
```handlebars
{{date ROW.A "DD/MM/YYYY"}}  {{predictAccount ROW.B ROW.C prefix="Expenses:"}}  {{amount ROW.D}}
    Assets:Checking:SBI                           {{negate (amount ROW.D)}}
```

**After:**
```handlebars
{{date ROW.A "DD/MM/YYYY"}}  {{predictAccountWithRules ROW.B ROW.C prefix="Expenses:"}}  {{amount ROW.D}}
    Assets:Checking:SBI                           {{negate (amount ROW.D)}}
```

### 3. How It Works

1. `predictAccountWithRules` first tries to match transaction data against your regex rules
2. If a rule matches, it returns the configured account
3. If no rules match, it falls back to the original TF-IDF based prediction
4. Rules are processed in order - first match wins
5. Rules respect the `prefix` parameter if provided

## Example Rules

### E-commerce
- **Pattern**: `AMAZON.*|AMZ.*` → **Account**: `Expenses:Shopping:Online`
- **Pattern**: `FLIPKART|FKRT.*` → **Account**: `Expenses:Shopping:Online`

### Transportation  
- **Pattern**: `UBER|LYFT|OLA` → **Account**: `Expenses:Transportation:Rideshare`
- **Pattern**: `METRO|SUBWAY` → **Account**: `Expenses:Transportation:Public`

### Utilities
- **Pattern**: `ELECTRIC.*|POWER.*` → **Account**: `Expenses:Utilities:Electric`
- **Pattern**: `WATER.*|H2O` → **Account**: `Expenses:Utilities:Water`

### Banking
- **Pattern**: `ATM.*FEE|WITHDRAWAL.*FEE` → **Account**: `Expenses:Banking:Fees`
- **Pattern**: `INTEREST.*CREDIT` → **Account**: `Income:Interest:Bank`

## Technical Details

### Configuration Storage
Rules are stored in the `paisa.yaml` configuration file under `account_rules`:

```yaml
account_rules:
  - name: "Amazon Purchases"
    pattern: "AMAZON.*"
    account: "Expenses:Shopping:Online" 
    description: "Amazon purchases"
    enabled: true
```

### API Endpoints
- `GET /api/account-rules` - List all rules
- `POST /api/account-rules/upsert` - Create/update a rule
- `POST /api/account-rules/delete` - Delete a rule

### Function Signature
```javascript
predictAccountWithRules(...args, options)
```

Same parameters as `predictAccount`, with the added regex matching logic.

## Benefits

1. **Precise Matching**: Regex patterns allow exact matching of transaction descriptions
2. **Fallback Support**: Falls back to ML-based prediction when no rules match  
3. **Easy Management**: Web UI for creating and managing rules
4. **Backward Compatible**: Existing `predictAccount` templates continue to work
5. **Configurable**: Rules can be enabled/disabled without deletion