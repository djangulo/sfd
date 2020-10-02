let user = Cypress.env("BASICAUTH_USER") ? Cypress.env("BASICAUTH_USER") : "";
let pass = Cypress.env("BASICAUTH_PASS") ? Cypress.env("BASICAUTH_PASS") : "";

describe("Item endpoint tests", function () {
  it("contains a list of items", () => {
    cy.visit("/", { auth: { username: user, password: pass } });
  });
  // Can navigate to item list
  it(`should navigate to item list from the nav`, () => {
    cy.get("nav").within(() => {
      cy.get('a[href="/articulos"]').click();
    });
  });

  it(`should contain the word "articulos" on the url'`, () => {
    cy.url().should("include", "/articulos");
  });

  it(`item list should contain pagination and at least some items`, () => {
    cy.get("div.pagination").should("exist");
    cy.contains("Test Item");
  });
  it("can see details of an item", () => {
    cy.get("h4.header").first().click();
    cy.url().should("include", "/articulos/test-item");
    cy.contains("Test Item");
  });
});
