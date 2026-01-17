import { test, expect } from '@playwright/test';

/**
 * E2E tests for user registration flow
 * Tests the RegisterView component and registration functionality
 */

test.describe('User Registration', () => {
  // Base URL for the frontend application
  const BASE_URL = 'http://localhost:5173';
  
  test.beforeEach(async ({ page }) => {
    // Navigate to the registration page before each test
    await page.goto(`${BASE_URL}/register`);
  });

  test('should display registration form with all required fields', async ({ page }) => {
    // Verify page title
    await expect(page.getByRole('heading', { name: 'Stwórz konto' })).toBeVisible();
    
    // Verify description text
    await expect(page.getByText('Dołącz do Ask Your Feed i zacznij zadawać pytania swojemu feedowi X')).toBeVisible();
    
    // Verify all form fields are present
    await expect(page.getByLabel('Adres e-mail')).toBeVisible();
    await expect(page.getByLabel('Hasło', { exact: true })).toBeVisible();
    await expect(page.getByLabel('Potwierdź hasło')).toBeVisible();
    await expect(page.getByLabel('Nazwa użytkownika X (Twitter)')).toBeVisible();
    
    // Verify submit button
    await expect(page.getByRole('button', { name: 'Zarejestruj się' })).toBeVisible();
    
    // Verify login link
    await expect(page.getByRole('link', { name: 'Zaloguj się' })).toBeVisible();
  });

  test('should show validation errors for empty form submission', async ({ page }) => {
    // Click submit button without filling any fields
    await page.getByRole('button', { name: 'Zarejestruj się' }).click();
    
    // Wait for validation errors to appear
    // Note: The exact error messages depend on your validation schema
    await expect(page.locator('text=/wymagane|required/i')).toBeVisible();
  });

  test('should show error when passwords do not match', async ({ page }) => {
    // Fill in the form with mismatched passwords
    await page.getByLabel('Adres e-mail').fill('test@example.com');
    await page.getByLabel('Hasło', { exact: true }).fill('Password123!');
    await page.getByLabel('Potwierdź hasło').fill('DifferentPassword123!');
    await page.getByLabel('Nazwa użytkownika X (Twitter)').fill('testuser');
    
    // Submit the form
    await page.getByRole('button', { name: 'Zarejestruj się' }).click();
    
    // Wait for password mismatch error
    await expect(page.locator('text=/hasła.*zgodne|passwords.*match/i')).toBeVisible();
  });

  test('should successfully register a new user with valid data', async ({ page }) => {
    // Generate unique email to avoid conflicts
    const timestamp = Date.now();
    const testEmail = `test${timestamp}@example.com`;
    const testPassword = 'SecurePassword123!';
    const testUsername = `dev1047600`;
    
    // Fill in the registration form
    await page.getByLabel('Adres e-mail').fill(testEmail);
    await page.getByLabel('Hasło', { exact: true }).fill(testPassword);
    await page.getByLabel('Potwierdź hasło').fill(testPassword);
    await page.getByLabel('Nazwa użytkownika X (Twitter)').fill(testUsername);
    
    // Submit the form
    await page.getByRole('button', { name: 'Zarejestruj się' }).click();
    
    // Wait for the button to show loading state
    await expect(page.getByRole('button', { name: 'Rejestrowanie...' })).toBeVisible();
    
    // After successful registration, user should be redirected to dashboard
    // Wait for navigation to complete (with timeout for API call)
    await page.waitForURL(`${BASE_URL}/`, { timeout: 10000 });
    
    // Verify we're on the dashboard page
    expect(page.url()).toBe(`${BASE_URL}/`);
    
    // Verify session token is stored in localStorage
    const sessionToken = await page.evaluate(() => localStorage.getItem('session_token'));
    expect(sessionToken).toBeTruthy();
    
    // Verify user data is stored in localStorage
    const userData = await page.evaluate(() => localStorage.getItem('user'));
    expect(userData).toBeTruthy();
    
    const user = JSON.parse(userData!);
    expect(user.email).toBe(testEmail);
    expect(user.x_username).toBe(testUsername);
  });

  test('should show error when email is already registered', async ({ page }) => {
    // Use an email that's likely already registered
    // Note: This test assumes you have a test user in your database
    const existingEmail = 'existing@example.com';
    
    await page.getByLabel('Adres e-mail').fill(existingEmail);
    await page.getByLabel('Hasło', { exact: true }).fill('Password123!');
    await page.getByLabel('Potwierdź hasło').fill('Password123!');
    await page.getByLabel('Nazwa użytkownika X (Twitter)').fill('testuser');
    
    // Submit the form
    await page.getByRole('button', { name: 'Zarejestruj się' }).click();
    
    // Wait for error message about duplicate email
    await expect(page.getByText(/Email jest już zarejestrowany/i)).toBeVisible({ timeout: 10000 });
  });

  test('should navigate to login page when clicking login link', async ({ page }) => {
    // Click the login link
    await page.getByRole('link', { name: 'Zaloguj się' }).click();
    
    // Verify navigation to login page
    await page.waitForURL(`${BASE_URL}/login`);
    expect(page.url()).toBe(`${BASE_URL}/login`);
  });

  test('should handle form input correctly', async ({ page }) => {
    const testEmail = 'test@example.com';
    const testPassword = 'Password123!';
    const testUsername = 'testuser';
    
    // Fill in each field and verify the value
    const emailInput = page.getByLabel('Adres e-mail');
    await emailInput.fill(testEmail);
    await expect(emailInput).toHaveValue(testEmail);
    
    const passwordInput = page.getByLabel('Hasło', { exact: true });
    await passwordInput.fill(testPassword);
    await expect(passwordInput).toHaveValue(testPassword);
    
    const confirmPasswordInput = page.getByLabel('Potwierdź hasło');
    await confirmPasswordInput.fill(testPassword);
    await expect(confirmPasswordInput).toHaveValue(testPassword);
    
    const usernameInput = page.getByLabel('Nazwa użytkownika X (Twitter)');
    await usernameInput.fill(testUsername);
    await expect(usernameInput).toHaveValue(testUsername);
  });

  test('should have proper accessibility attributes', async ({ page }) => {
    // Verify form has proper structure
    const form = page.locator('form');
    await expect(form).toBeVisible();
    
    // Verify all inputs have associated labels
    await expect(page.getByLabel('Adres e-mail')).toHaveAttribute('type', 'email');
    await expect(page.getByLabel('Hasło', { exact: true })).toHaveAttribute('type', 'password');
    await expect(page.getByLabel('Potwierdź hasło')).toHaveAttribute('type', 'password');
    await expect(page.getByLabel('Nazwa użytkownika X (Twitter)')).toHaveAttribute('type', 'text');
    
    // Verify autocomplete attributes
    await expect(page.getByLabel('Adres e-mail')).toHaveAttribute('autocomplete', 'email');
    await expect(page.getByLabel('Hasło', { exact: true })).toHaveAttribute('autocomplete', 'new-password');
    await expect(page.getByLabel('Potwierdź hasło')).toHaveAttribute('autocomplete', 'new-password');
  });
});
