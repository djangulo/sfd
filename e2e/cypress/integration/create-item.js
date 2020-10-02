let user = Cypress.env("BASICAUTH_USER") ? Cypress.env("BASICAUTH_USER") : "";
let pass = Cypress.env("BASICAUTH_PASS") ? Cypress.env("BASICAUTH_PASS") : "";
describe("Create an item", function () {
  before(() => {
    cy.clearCookies();
  });
  beforeEach(() => {
    Cypress.Cookies.preserveOnce("sfd-session-id");
  });
  it(`navigating to articles/new immediately brings the login screen`, () => {
    cy.visit("/articulos/nuevo", {
      auth: { username: user, password: pass },
    });
    cy.location("pathname").should("eq", "/iniciar-sesion");
    cy.location("search").should("contain", "next");
  });
  it(`should redirect to new item form once logged in`, () => {
    cy.get(`input[name="username"]`).clear();
    cy.get(`input[name="password"]`).clear();
    cy.get(`input[name="username"]`).type("testuser-00");
    cy.get(`input[name="password"]`).type(`testuser-00`);
    cy.get(`button[type="submit"]`).click();
    cy.location("pathname").should("eq", "/articulos/nuevo");
  });

  // it(`once logged in, own profile should be visible`, () => {
  //   cy.login("testuser-00", "testuser-00");
  //   cy.visit("/cuentas/testuser-00", {
  //     auth: { username: user, password: pass },
  //   });
  // });

  // it(`clicking the empty form returns errors`, () => {
  //   cy.get(`button[type="submit"]`).click();
  //   cy.get(`.ui.error.message`).contains("no puede estar vacío");
  // });
  // it(`enters some bogus data`, () => {
  //   cy.get(`input[name="username"]`).type("madeupusername");
  //   cy.get(`input[name="password"]`).type(`madeuppassowrd`);
  //   cy.get(`button[type="submit"]`).click();
  //   cy.get(`.ui.error.message`).contains(
  //     "nombre de usuario o contraseña inválido"
  //   );
  // });
  // it(`has the session cookie "sfd-session-id"`, () => {
  //   cy.getCookie("sfd-session-id").should("exist");
  // });
  // it(`should be able to logout`, () => {
  //   cy.get("nav").within(() => {
  //     cy.get(`i.dropdown.icon`).click();
  //     cy.get(`a[href="/cerrar-sesion"]`).click();
  //   });
  // });
  // it(`session cookie "sfd-session-id" should not exist`, () => {
  //   cy.getCookie("sfd-session-id").should("not.exist");
  // });
  // it(`should redirect to "/"`, () => {
  //   cy.location("pathname").should("eq", "/");
  // });
});
