# Roadmap

Some features I would like to implement in the coming months.

## Phase 1: Core User Experience (High Priority)
*Focus: Completing the essential user loop and personalization.*

### 1. Money Handling Refactor
- [ ] **Multi-currency Support:** Replace current money implementation with a robust external library (e.g., `rhymond/go-money` or `govalues/decimal`) to handle multiple currencies and exchange rates correctly.
- [ ] **Migration:** Update existing monetary values in the database to be compatible with the new library structure if necessary.

### 2. User Profile & Settings
- [ ] **Global Configuration:** Add `DISABLE_REGISTRATION` env var to toggle public sign-ups.
- [ ] **User Entity Update:** Add `Currency` and `Locale` fields to the `User` entity (Database Migration required).
- [ ] **Profile Page:** Create a settings page where users can:
    - Change their display name/email.
    - Set their preferred currency (overriding global default).
    - Change password.

### 2. Feedback System (Toast Notifications)
- [x] **Scaffolding:** Basic Templ component and Handler logic exist.
- [x] **Integration:** Ensure seamless bridge between Server-Side `HX-Trigger` events and Client-Side Alpine.js toast manager.
- [x] **Standardization:** Replace all ad-hoc alerts/redirects with standardized Toast feedback for:
    - Successful saves/updates.
    - Form validation errors.
    - System errors.

### 3. Basic Reporting
- [ ] **Monthly Overview:** Dashboard widget showing total Income vs. Expenses vs. Balance for the current month.
- [ ] **Category Breakdown:** Simple list or bar chart showing top spending categories.
- [ ] **Month-over-Month:** Basic comparison to the previous month (e.g., "You spent 10% less than last month").

---

## Phase 2: Enhanced Functionality (Medium Priority)
*Focus: Giving users more control and insight into their data.*

### 4. Data Management
- [ ] **Search & Filter:** Add search bar to Expense/Income lists (filter by date range, category, amount, description).
- [ ] **Pagination:** Implement efficient database pagination for large transaction lists.
- [ ] **Data Export:** functionality to export all data to CSV/JSON format for user backup.

### 5. Budgeting Features
- [ ] **Category Limits:** Allow users to set a maximum budget per category (e.g., "Groceries: $500/month").
- [ ] **Visual Indicators:** Progress bars on the dashboard showing budget usage (Green/Yellow/Red).
- [ ] **Balance Display:** Display total amount budgeted vs current balance for the month.

### 6. Recurring Transactions
- [ ] **Subscriptions:** Mark expenses as "Recurring" (Monthly/Yearly).
- [ ] **Auto-Entry:** *Optional:* Background job to automatically create recurring records (or just prompt the user).

---

## Phase 3: UX & Technical Polish (Low Priority)
*Focus: Visual appeal, security, and developer experience.*

### 7. User Interface
- [ ] **Dark Mode:** System-aware theme toggle using Tailwind's `dark:` modifier.
- [ ] **Mobile Optimization:** Polish touch targets and layout for the "Add Expense" flow on mobile devices.

### 8. Security & Ops
- [ ] **Rate Limiting:** Implement middleware to prevent abuse of Auth endpoints.
- [ ] **HSTS:** Enable Strict-Transport-Security for production builds.
- [ ] **Audit Logs:** Track sensitive actions (login, password change, data export). Maybe event sourcing?

### 9. Developer Experience
- [ ] **API:** Add a versioned REST API for core functionalities (CRUD for Expenses/Incomes).

### 9. Backup & Recovery
- [ ] **Automated Backups:** Add support for litestream to automate SQLite backups to S3/Azure Blob Storage.
